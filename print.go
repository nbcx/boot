package boot

import (
	"fmt"
	"io"
	"os"
)

type Print struct {
	// inReader is a reader defined by the user that replaces stdin
	inReader io.Reader
	// outWriter is a writer defined by the user that replaces stdout
	outWriter io.Writer
	// errWriter is a writer defined by the user that replaces stderr
	errWriter io.Writer
}

var log = &Print{}

// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
// Deprecated: Use SetOut and/or SetErr instead
func (c *Print) SetOutput(output io.Writer) {
	c.outWriter = output
	c.errWriter = output
}

// SetOut sets the destination for usage messages.
// If newOut is nil, os.Stdout is used.
func (c *Print) SetOut(newOut io.Writer) {
	c.outWriter = newOut
}

func (c *Print) GetOut() io.Writer {
	return c.outWriter
}

// SetErr sets the destination for error messages.
// If newErr is nil, os.Stderr is used.
func (c *Print) SetErr(newErr io.Writer) {
	c.errWriter = newErr
}

// SetIn sets the source for input data
// If newIn is nil, os.Stdin is used.
func (c *Print) SetIn(newIn io.Reader) {
	c.inReader = newIn
}

func (c *Print) GetIn() io.Reader {
	return c.inReader
}

// OutOrStdout returns output to stdout.
func (c *Print) OutOrStdout() io.Writer {
	return c.getOut(os.Stdout)
}

// OutOrStderr returns output to stderr
func (c *Print) OutOrStderr() io.Writer {
	return c.getOut(os.Stderr)
}

// ErrOrStderr returns output to stderr
func (c *Print) ErrOrStderr() io.Writer {
	return c.getErr(os.Stderr)
}

// InOrStdin returns input to stdin
func (c *Print) InOrStdin() io.Reader {
	return c.getIn(os.Stdin)
}

func (c *Print) getOut(def io.Writer) io.Writer {
	if c.outWriter == nil {
		c.outWriter = def
	}
	return c.outWriter
}

func (c *Print) getErr(def io.Writer) io.Writer {
	if c.errWriter == nil {
		c.errWriter = def
	}
	return c.errWriter
}

func (c *Print) getIn(def io.Reader) io.Reader {
	if c.inReader == nil {
		c.inReader = def
	}
	return c.inReader
}

// Print is a convenience method to Print to the defined output, fallback to Stderr if not set.
func (c *Print) Print(i ...interface{}) {
	fmt.Fprint(c.OutOrStderr(), i...)
}

// Println is a convenience method to Println to the defined output, fallback to Stderr if not set.
func (c *Print) Println(i ...interface{}) {
	c.Print(fmt.Sprintln(i...))
}

// Printf is a convenience method to Printf to the defined output, fallback to Stderr if not set.
func (c *Print) Printf(format string, i ...interface{}) {
	c.Print(fmt.Sprintf(format, i...))
}

// PrintErr is a convenience method to Print to the defined Err output, fallback to Stderr if not set.
func (c *Print) PrintErr(i ...interface{}) {
	fmt.Fprint(c.ErrOrStderr(), i...)
}

// PrintErrLn is a convenience method to Println to the defined Err output, fallback to Stderr if not set.
func (c *Print) PrintErrLn(i ...interface{}) {
	c.PrintErr(fmt.Sprintln(i...))
}

// PrintErrF is a convenience method to Printf to the defined Err output, fallback to Stderr if not set.
func (c *Print) PrintErrF(format string, i ...interface{}) {
	c.PrintErr(fmt.Sprintf(format, i...))
}
