package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/katexochen/ttrpcurl"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:   "ttrpcurl [flags] <socket> <method>",
		Short: "Make ttrpc calls based on a proto file",
		Args: cobra.MatchAll(
			cobra.ExactArgs(2),
		),
		RunE: runRoot,
	}

	rootCmd.Flags().StringP("data", "d", "", prettify(`
		Data for request contents. If the value is '@' then the request contents
		are read from stdin. For calls that accept a stream of requests, the
		contents should include all such request messages concatenated together
		(possibly delimited; see -format).`))
	rootCmd.Flags().String("format", "json", prettify(`
		The format of request data. The allowed values are 'json' or 'text'. For
		'json', the input data must be in JSON format. Multiple request values
		may be concatenated (messages with a JSON representation other than
		object must be separated by whitespace, such as a newline). For 'text',
		the input data must be in the protobuf text format, in which case
		multiple request values must be separated by the "record separator"
		ASCII character: 0x1E. The stream should not end in a record separator.
		If it does, it will be interpreted as a final, blank message after the
		separator.`))
	rootCmd.Flags().StringSlice("proto", []string{}, prettify(`
		The name of a proto source file. Source files given will be used to
		determine the RPC schema instead of querying for it from the remote
		server via the gRPC reflection API. When set: the 'list' action lists
		the services found in the given files and their imports (vs. those
		exposed by the remote server), and the 'describe' action describes
		symbols found in the given files. May specify more than one via multiple
		-proto flags. Imports will be resolved using the given -import-path
		flags. Multiple proto files can be specified by specifying multiple
		-proto flags. It is an error to use both -protoset and -proto flags.`))
	rootCmd.Flags().Bool("allow-unknown-fields", false, prettify(`
		When true, the request contents, if 'json' format is used, allows
		unknown fields to be present. They will be ignored when parsing
		the request.`))
	rootCmd.Flags().Duration("connect-timeout", 0, prettify(`
		The maximum time, in seconds, to wait for connection to be established.
		Defaults to 10 seconds.`))
	rootCmd.Flags().Bool("format-error", false, prettify(`
		When a non-zero status is returned, format the response using the
		value set by the -format flag .`))
	rootCmd.Flags().Duration("max-time", 0, prettify(`
		The maximum total time the operation can take, in seconds. This is
		useful for preventing batch jobs that use grpcurl from hanging due to
		slow or bad network links or due to incorrect stream method usage.`))
	rootCmd.Flags().Uint("max-msg-sz", 4194304, prettify(`
		The maximum encoded size of a response message, in bytes, that grpcurl
		will accept. Defaults 4 MiB.`))
	rootCmd.Flags().Bool("emit-defaults", false, prettify(`
		Emit default values for JSON-encoded responses.`))
	// rootCmd.Flags().Bool("msg-template", false, prettify(`
	// 	When describing messages, show a template of input data.`))

	// Unused flags, might be implemented in the future
	rootCmd.Flags().StringSlice("protoset", nil, "")
	rootCmd.Flags().MarkHidden("protoset")
	rootCmd.Flags().StringSlice("import-path", nil, "")
	rootCmd.Flags().MarkHidden("import-path")
	rootCmd.Flags().Bool("use-reflection", false, "")
	rootCmd.Flags().MarkHidden("use-reflection")
	rootCmd.Flags().StringP("add-header", "H", "", "")
	rootCmd.Flags().MarkHidden("add-headers")
	rootCmd.Flags().String("rpc-header", "", "")
	rootCmd.Flags().MarkHidden("rpc-header")
	rootCmd.Flags().String("reflect-header", "", "")
	rootCmd.Flags().MarkHidden("reflect-header")
	rootCmd.Flags().Bool("expand-headers", false, "")
	rootCmd.Flags().MarkHidden("expand-headers")
	rootCmd.Flags().String("protoset-out", "", "")
	rootCmd.Flags().MarkHidden("protoset-out")
	rootCmd.Flags().Bool("reflection", false, "")
	rootCmd.Flags().MarkHidden("reflection")

	// Unused flags for compability with grpcurl
	rootCmd.Flags().Bool("plaintext", false, "")
	rootCmd.Flags().MarkHidden("plaintext")
	rootCmd.Flags().Bool("insecure", false, "")
	rootCmd.Flags().MarkHidden("insecure")
	rootCmd.Flags().Bool("cacert", false, "")
	rootCmd.Flags().MarkHidden("cacert")
	rootCmd.Flags().Bool("cert", false, "")
	rootCmd.Flags().MarkHidden("cert")
	rootCmd.Flags().Bool("key", false, "")
	rootCmd.Flags().MarkHidden("key")
	rootCmd.Flags().String("authority", "", "")
	rootCmd.Flags().MarkHidden("authority")
	rootCmd.Flags().String("user-agent", "", "")
	rootCmd.Flags().MarkHidden("user-agent")
	rootCmd.Flags().Duration("keepalive-time", 0, "")
	rootCmd.Flags().MarkHidden("keepalive-time")
	rootCmd.Flags().String("servername", "", "")
	rootCmd.Flags().MarkHidden("servername")

	return rootCmd
}

func runRoot(cmd *cobra.Command, args []string) error {
	flags, err := parseRootFlags(cmd)
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if err := warnRootCompabilityFlags(cmd); err != nil {
		return fmt.Errorf("parsing compability flags: %w", err)
	}

	var data []byte
	if flags.data == "@" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read data from stdin: %w", err)
		}
	} else {
		data = []byte(flags.data)
	}

	return ttrpcurl.Execute(flags.proto, args[0], args[1], data)
}

type rootFlags struct {
	verbose            bool // persistent
	data               string
	format             string
	proto              []string
	allowUnknownFields bool
	connectTimeout     time.Duration
	formatError        bool
	maxTime            time.Duration
	maxMsgSz           uint
	emitDefaults       bool
}

func parseRootFlags(cmd *cobra.Command) (*rootFlags, error) {
	f := &rootFlags{}

	var err error
	f.verbose, err = cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}
	f.data, err = cmd.Flags().GetString("data")
	if err != nil {
		return nil, err
	}
	f.format, err = cmd.Flags().GetString("format")
	if err != nil {
		return nil, err
	}
	f.proto, err = cmd.Flags().GetStringSlice("proto")
	if err != nil {
		return nil, err
	}
	f.allowUnknownFields, err = cmd.Flags().GetBool("allow-unknown-fields")
	if err != nil {
		return nil, err
	}
	f.connectTimeout, err = cmd.Flags().GetDuration("connect-timeout")
	if err != nil {
		return nil, err
	}
	f.formatError, err = cmd.Flags().GetBool("format-error")
	if err != nil {
		return nil, err
	}
	f.maxTime, err = cmd.Flags().GetDuration("max-time")
	if err != nil {
		return nil, err
	}
	f.maxMsgSz, err = cmd.Flags().GetUint("max-msg-sz")
	if err != nil {
		return nil, err
	}
	f.emitDefaults, err = cmd.Flags().GetBool("emit-defaults")
	if err != nil {
		return nil, err
	}

	if f.emitDefaults && f.format != "json" {
		return nil, fmt.Errorf("flag --emit-defaults is only supported for --format=json")
	}

	return f, nil
}

func warnRootCompabilityFlags(cmd *cobra.Command) error {
	compFlags := []struct {
		name    string
		warning string
	}{
		{"plaintext", "The flag deactivates TLS in grpcurl, but ttrpcurl communicates over a unix domain socket. It never uses TLS."},
		{"insecure", ""},
		{"cacert", ""},
		{"cert", ""},
		{"key", ""},
		{"authority", ""},
		{"user-agent", ""},
		{"keepalive-time", ""},
		{"servername", ""},
	}

	for _, flag := range compFlags {
		if cmd.Flags().Changed(flag.name) {
			fmt.Printf("WARN: flag %s is unused and only provided for compability with grpcurl. %s\n", flag.name, flag.warning)
		}
	}

	return nil
}
