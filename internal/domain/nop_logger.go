package domain

import "context"

type nopLogger struct{}

func (n *nopLogger) DebugContext(ctx context.Context, msg string, args ...any) {}
func (n *nopLogger) InfoContext(ctx context.Context, msg string, args ...any)  {}
func (n *nopLogger) WarnContext(ctx context.Context, msg string, args ...any)  {}
func (n *nopLogger) ErrorContext(ctx context.Context, msg string, args ...any) {}
func (n *nopLogger) With(args ...any) Logger                                   { return n }
func (n *nopLogger) WithGroup(name string) Logger                              { return n }
