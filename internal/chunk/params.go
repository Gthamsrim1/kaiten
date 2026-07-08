package chunk

import "fmt"

type Params struct {
	MinSize int
	AvgSize int
	MaxSize int
	Window  int
}

var Default = Params{
	MinSize: 16 * 1024,
	AvgSize: 64 * 1024,
	MaxSize: 256 * 1024,
	Window:  64,
}

func (p Params) Validate() error {
	if p.Window <= 0 {
		return fmt.Errorf("window must be greater than 0")
	}

	if p.MinSize <= 0 {
		return fmt.Errorf("minimum chunk size must be greater than 0")
	}

	if p.AvgSize <= 0 {
		return fmt.Errorf("average chunk size must be greater than 0")
	}

	if p.MaxSize <= 0 {
		return fmt.Errorf("maximum chunk size must be greater than 0")
	}

	if p.MinSize > p.AvgSize {
		return fmt.Errorf("minimum chunk size cannot exceed average chunk size")
	}

	if p.AvgSize > p.MaxSize {
		return fmt.Errorf("average chunk size cannot exceed maximum chunk size")
	}

	if p.AvgSize&(p.AvgSize-1) != 0 {
		return fmt.Errorf("average chunk size must be a power of two")
	}

	if p.Window > p.MinSize {
		return fmt.Errorf("window size cannot exceed minimum chunk size")
	}

	return nil
}
