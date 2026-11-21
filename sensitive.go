package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/option"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	sensitiveRuleID = "SENSITIVE_FIELD_UNREACHABLE"

	sensitiveOptionKey          = "sensitive_option"
	sensitiveRootSuffixesKey    = "sensitive_root_message_suffixes"
	sensitiveRootPackagesKey    = "sensitive_root_service_packages"
	sensitiveAllowedExternalKey = "sensitive_allowed_external_types"
)

var sensitiveRule = &check.RuleSpec{
	ID:      sensitiveRuleID,
	Purpose: "Checks that no field marked with the sensitive option, and no external message type, is reachable from root messages and RPCs.",
	Type:    check.RuleTypeLint,
	Handler: check.RuleHandlerFunc(checkSensitive),
}

type sensitiveConfig struct {
	ext              protoreflect.ExtensionType
	unmarshal        proto.UnmarshalOptions
	rootSuffixes     []string
	rootPackages     []string
	allowedExternals []string
}

func checkSensitive(_ context.Context, w check.ResponseWriter, req check.Request) error {
	files := new(protoregistry.Files)
	imports := map[string]bool{}
	for _, f := range req.FileDescriptors() {
		fd := f.ProtoreflectFileDescriptor()
		if err := files.RegisterFile(fd); err != nil {
			return fmt.Errorf("%s: %w", sensitiveRuleID, err)
		}
		if f.IsImport() {
			imports[fd.Path()] = true
		}
	}
	cfg, err := readSensitiveConfig(req.Options(), files)
	if err != nil {
		return err
	}
	walker := &sensitiveWalker{cfg: cfg, imports: imports, marked: map[protoreflect.FullName]bool{}}
	for _, f := range req.FileDescriptors() {
		if f.IsImport() {
			continue
		}
		fd := f.ProtoreflectFileDescriptor()
		err := eachMessage(fd.Messages(), func(m protoreflect.MessageDescriptor) error {
			return walker.checkRootMessage(w, m)
		})
		if err != nil {
			return err
		}
		if !slices.Contains(cfg.rootPackages, string(fd.Package())) {
			continue
		}
		services := fd.Services()
		for i := range services.Len() {
			methods := services.Get(i).Methods()
			for j := range methods.Len() {
				if err := walker.checkRootMethod(w, methods.Get(j)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func readSensitiveConfig(opts option.Options, files *protoregistry.Files) (*sensitiveConfig, error) {
	name, err := option.GetStringValue(opts, sensitiveOptionKey)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("%s needs the plugin option %q, the full name of a bool field option", sensitiveRuleID, sensitiveOptionKey)
	}
	d, err := files.FindDescriptorByName(protoreflect.FullName(name))
	if err != nil {
		return nil, fmt.Errorf("%s %q: %w", sensitiveOptionKey, name, err)
	}
	xd, ok := d.(protoreflect.ExtensionDescriptor)
	if !ok || xd.ContainingMessage().FullName() != "google.protobuf.FieldOptions" || xd.Kind() != protoreflect.BoolKind || xd.IsList() {
		return nil, fmt.Errorf("%s %q is not a bool extension of google.protobuf.FieldOptions", sensitiveOptionKey, name)
	}
	cfg := &sensitiveConfig{ext: dynamicpb.NewExtensionType(xd)}
	types := new(protoregistry.Types)
	if err := types.RegisterExtension(cfg.ext); err != nil {
		return nil, err
	}
	cfg.unmarshal = proto.UnmarshalOptions{Resolver: types}
	if cfg.rootSuffixes, err = option.GetStringSliceValue(opts, sensitiveRootSuffixesKey); err != nil {
		return nil, err
	}
	if cfg.rootPackages, err = option.GetStringSliceValue(opts, sensitiveRootPackagesKey); err != nil {
		return nil, err
	}
	if cfg.allowedExternals, err = option.GetStringSliceValue(opts, sensitiveAllowedExternalKey); err != nil {
		return nil, err
	}
	if len(cfg.rootSuffixes) == 0 && len(cfg.rootPackages) == 0 {
		return nil, fmt.Errorf("%s needs %q or %q", sensitiveRuleID, sensitiveRootSuffixesKey, sensitiveRootPackagesKey)
	}
	return cfg, nil
}

type sensitiveWalker struct {
	cfg     *sensitiveConfig
	imports map[string]bool
	marked  map[protoreflect.FullName]bool
}

func (s *sensitiveWalker) checkRootMessage(w check.ResponseWriter, m protoreflect.MessageDescriptor) error {
	if !hasAnySuffix(string(m.Name()), s.cfg.rootSuffixes) {
		return nil
	}
	fields := m.Fields()
	for i := range fields.Len() {
		first := fields.Get(i)
		report := func(path string) {
			w.AddAnnotation(check.WithMessagef("%s reaches %s.", m.FullName(), path), check.WithDescriptor(first))
		}
		if err := s.field(first, string(m.FullName()), map[protoreflect.FullName]bool{m.FullName(): true}, report); err != nil {
			return err
		}
	}
	return nil
}

func (s *sensitiveWalker) checkRootMethod(w check.ResponseWriter, method protoreflect.MethodDescriptor) error {
	for _, side := range []struct {
		name string
		msg  protoreflect.MessageDescriptor
	}{{"request", method.Input()}, {"response", method.Output()}} {
		if s.isExternal(side.msg) {
			if !s.isAllowed(side.msg) {
				w.AddAnnotation(check.WithMessagef("%s %s is external type %s.", method.FullName(), side.name, side.msg.FullName()), check.WithDescriptor(method))
			}
			continue
		}
		report := func(path string) {
			w.AddAnnotation(check.WithMessagef("%s %s reaches %s.", method.FullName(), side.name, path), check.WithDescriptor(method))
		}
		if err := s.message(side.msg, string(side.msg.FullName()), map[protoreflect.FullName]bool{}, report); err != nil {
			return err
		}
	}
	return nil
}

func (s *sensitiveWalker) message(m protoreflect.MessageDescriptor, path string, onPath map[protoreflect.FullName]bool, report func(string)) error {
	if onPath[m.FullName()] {
		return nil
	}
	onPath[m.FullName()] = true
	defer delete(onPath, m.FullName())
	fields := m.Fields()
	for i := range fields.Len() {
		if err := s.field(fields.Get(i), path, onPath, report); err != nil {
			return err
		}
	}
	return nil
}

// field calls report with the path of each sensitive field or disallowed
// external type reachable through f. path is the dotted path to the message
// holding f, and onPath holds the messages on that path, which are not entered
// again.
func (s *sensitiveWalker) field(f protoreflect.FieldDescriptor, path string, onPath map[protoreflect.FullName]bool, report func(string)) error {
	at := path + "." + string(f.Name())
	marked, err := s.isMarked(f)
	if err != nil {
		return err
	}
	if marked {
		report(at)
		return nil
	}
	next := f.Message()
	if next == nil {
		return nil
	}
	if s.isExternal(next) {
		if !s.isAllowed(next) {
			report(fmt.Sprintf("%s of external type %s", at, next.FullName()))
		}
		return nil
	}
	return s.message(next, at, onPath, report)
}

func (s *sensitiveWalker) isExternal(m protoreflect.MessageDescriptor) bool {
	return s.imports[m.ParentFile().Path()]
}

func (s *sensitiveWalker) isAllowed(m protoreflect.MessageDescriptor) bool {
	return slices.Contains(s.cfg.allowedExternals, string(m.FullName()))
}

func (s *sensitiveWalker) isMarked(f protoreflect.FieldDescriptor) (bool, error) {
	if v, ok := s.marked[f.FullName()]; ok {
		return v, nil
	}
	v, err := s.readMark(f)
	if err != nil {
		return false, fmt.Errorf("reading %s on %s: %w", s.cfg.ext.TypeDescriptor().FullName(), f.FullName(), err)
	}
	s.marked[f.FullName()] = v
	return v, nil
}

func (s *sensitiveWalker) readMark(f protoreflect.FieldDescriptor) (bool, error) {
	opts := f.Options()
	b, err := proto.Marshal(opts)
	if err != nil {
		return false, err
	}
	m := opts.ProtoReflect().New().Interface()
	if err := s.cfg.unmarshal.Unmarshal(b, m); err != nil {
		return false, err
	}
	return m.ProtoReflect().Get(s.cfg.ext.TypeDescriptor()).Bool(), nil
}

func eachMessage(ms protoreflect.MessageDescriptors, fn func(protoreflect.MessageDescriptor) error) error {
	for i := range ms.Len() {
		m := ms.Get(i)
		if m.IsMapEntry() {
			continue
		}
		if err := fn(m); err != nil {
			return err
		}
		if err := eachMessage(m.Messages(), fn); err != nil {
			return err
		}
	}
	return nil
}

func hasAnySuffix(name string, suffixes []string) bool {
	return slices.ContainsFunc(suffixes, func(s string) bool { return strings.HasSuffix(name, s) })
}
