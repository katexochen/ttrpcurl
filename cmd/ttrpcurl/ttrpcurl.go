package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/katexochen/ttrpcurl"
	"github.com/katexochen/ttrpcurl/proto"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func newRootCmd() *cobra.Command {
	cobra.EnableCommandSorting = false

	cmd := &cobra.Command{
		Use:   "ttrpcurl [flags] <socket> <method>",
		Short: "Make ttrpc calls based on a proto file",
		Args: cobra.MatchAll(
			cobra.ExactArgs(2),
		),
		RunE: runRoot,
	}

	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output.")
	cmd.PersistentFlags().StringSlice("proto", []string{}, prettify(`
		The path of a proto source file. May specify more than one via repeated
		use of the flag or by passing a comma separated list of strings.`))
	// Imports will be resolved using the given -import-path flags.
	// It is an error to use both -protoset and -proto flags.
	must(cmd.MarkPersistentFlagRequired("proto"))

	cmd.Flags().StringP("data", "d", "", prettify(`
		Data for request contents. If the value is '@' then the request contents
		are read from stdin. For calls that accept a stream of requests, the
		contents should include all such request messages concatenated together
		(possibly delimited; see -format).`))
	cmd.Flags().String("format", "json", prettify(`
		The format of request data. The allowed values are 'json' or 'text'. For
		'json', the input data must be in JSON format. Multiple request values
		may be concatenated (messages with a JSON representation other than
		object must be separated by whitespace, such as a newline). For 'text',
		the input data must be in the protobuf text format, in which case
		multiple request values must be separated by the "record separator"
		ASCII character: 0x1E. The stream should not end in a record separator.
		If it does, it will be interpreted as a final, blank message after the
		separator.`))
	cmd.Flags().Bool("allow-unknown-fields", false, prettify(`
		When true, the request contents, if 'json' format is used, allows
		unknown fields to be present. They will be ignored when parsing
		the request.`))
	cmd.Flags().Duration("connect-timeout", 0, prettify(`
		The maximum time, in seconds, to wait for connection to be established.
		Defaults to 10 seconds.`))
	cmd.Flags().Bool("format-error", false, prettify(`
		When a non-zero status is returned, format the response using the
		value set by the -format flag .`))
	cmd.Flags().Duration("max-time", 0, prettify(`
		The maximum total time the operation can take, in seconds. This is
		useful for preventing batch jobs that use grpcurl from hanging due to
		slow or bad network links or due to incorrect stream method usage.`))
	cmd.Flags().Uint("max-msg-sz", 4194304, prettify(`
		The maximum encoded size of a response message, in bytes, that grpcurl
		will accept. Defaults 4 MiB.`))
	cmd.Flags().Bool("emit-defaults", false, prettify(`
		Emit default values for JSON-encoded responses.`))

	// Unused flags, might be implemented in the future
	// rootCmd.Flags().StringSlice("protoset", nil, "")
	// rootCmd.Flags().StringSlice("import-path", nil, "")
	// rootCmd.Flags().Bool("use-reflection", false, "")
	// rootCmd.Flags().StringP("add-header", "H", "", "")
	// rootCmd.Flags().String("rpc-header", "", "")
	// rootCmd.Flags().String("reflect-header", "", "")
	// rootCmd.Flags().Bool("expand-headers", false, "")
	// rootCmd.Flags().String("protoset-out", "", "")
	// rootCmd.Flags().Bool("reflection", false, "")

	return cmd
}

func runRoot(cmd *cobra.Command, args []string) error {
	flags, err := parseRootFlags(cmd)
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	var data []byte
	if flags.data == "@" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read data from stdin: %w", err)
		}
	} else if flags.data != "" {
		data = []byte(flags.data)
	}

	parser := proto.NewParser()
	source, err := parser.ParseFiles(flags.proto...)
	if err != nil {
		return fmt.Errorf("parsing proto files: %w", err)
	}

	dialer := net.Dialer{}
	conn, err := dialer.Dial("unix", args[0])
	if err != nil {
		return fmt.Errorf("dialing unix domain socket: %w", err)
	}
	defer conn.Close()

	marshaler := &protojson.MarshalOptions{Multiline: true}
	client := ttrpcurl.NewClient(conn, source, marshaler)

	return client.Call(cmd.Context(), args[1], data)
}

type rootFlags struct {
	verbose            bool     // persistent
	proto              []string // persistent
	data               string
	format             string
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
