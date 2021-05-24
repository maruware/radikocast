package radikocast

import (
	"context"
	"fmt"
)

type Converter interface {
	Convert(ctx context.Context, input, output string) error
}

func NewConverter(format string) (Converter, error) {
	switch format {
	case AudioFormatM4A:
		return &ConverterM4A{}, nil
	case AudioFormatMP3:
		return &ConverterMP3{}, nil
	default:
		return nil, fmt.Errorf("Bad format: %s", format)
	}
}

type ConverterMP3 struct {
}

func (c *ConverterMP3) Convert(ctx context.Context, input, output string) error {
	f, err := newFfmpeg(ctx)
	if err != nil {
		return err
	}

	f.setInput(input)
	f.setArgs(
		"-c:a", "libmp3lame",
		"-q:a", "2",
		"-y", // overwrite the output file without asking
	)
	return f.run(output)
}

type ConverterM4A struct {
}

func (c *ConverterM4A) Convert(ctx context.Context, input, output string) error {
	f, err := newFfmpeg(ctx)
	if err != nil {
		return err
	}

	f.setInput(input)
	f.setArgs(
		"-c:a", "copy",
		"-y", // overwrite the output file without asking
	)
	return f.run(output)
}
