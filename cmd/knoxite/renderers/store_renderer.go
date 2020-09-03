/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package renderers

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/knoxite/knoxite"
	"github.com/muesli/goprogressbar"
)

type StoreRenderer struct {
	DefaultRenderer    knoxite.DefaultRenderer
	FileProgressBar    goprogressbar.ProgressBar
	OverallProgressBar goprogressbar.ProgressBar
	MultiProgressBar   goprogressbar.MultiProgressBar
	LastPath           string
	Items              int64
	DisposeMessage     string
}

func (r *StoreRenderer) Init() error {
	err := r.DefaultRenderer.Init()
	if err != nil {
		return err
	}
	startTime := time.Now()

	r.FileProgressBar = goprogressbar.ProgressBar{Width: 40}
	r.OverallProgressBar = goprogressbar.ProgressBar{
		Text:  fmt.Sprintf("%d of %d total", 0, 0),
		Width: 60,
		PrependTextFunc: func(p *goprogressbar.ProgressBar) string {
			return fmt.Sprintf("%s/s",
				knoxite.SizeToString(uint64(float64(p.Current)/time.Since(startTime).Seconds())))
		},
	}

	r.MultiProgressBar = goprogressbar.MultiProgressBar{}
	r.MultiProgressBar.AddProgressBar(&r.FileProgressBar)
	r.MultiProgressBar.AddProgressBar(&r.OverallProgressBar)

	r.Items = 1
	return nil
}

func (r *StoreRenderer) Render(p knoxite.Progress) error {
	if err := r.DefaultRenderer.Render(p); err != nil {
		return err
	}

	if p.Error != nil {
		fmt.Println()
		return p.Error
	}

	if p.Path != r.LastPath && r.LastPath != "" {
		r.Items++
		fmt.Println()
	}

	r.FileProgressBar.Total = int64(p.CurrentItemStats.Size)
	r.FileProgressBar.Current = int64(p.CurrentItemStats.Transferred)
	r.FileProgressBar.PrependText = fmt.Sprintf("%s  %s/s",
		knoxite.SizeToString(uint64(r.FileProgressBar.Current)),
		knoxite.SizeToString(p.TransferSpeed()))

	r.OverallProgressBar.Total = int64(p.TotalStatistics.Size)
	r.OverallProgressBar.Current = int64(p.TotalStatistics.Transferred)
	r.OverallProgressBar.Text = fmt.Sprintf("%s / %s (%s of %s)",
		knoxite.SizeToString(uint64(r.OverallProgressBar.Current)),
		knoxite.SizeToString(uint64(r.OverallProgressBar.Total)),
		humanize.Comma(r.Items),
		humanize.Comma(int64(p.TotalStatistics.Files+p.TotalStatistics.Dirs+p.TotalStatistics.SymLinks)))

	if p.Path != r.LastPath {
		r.LastPath = p.Path
		r.FileProgressBar.Text = p.Path
	}

	r.MultiProgressBar.LazyPrint()

	return nil
}

func (r *StoreRenderer) Close() error {
	if err := r.DefaultRenderer.Close(); err != nil {
		return err
	}
	fmt.Println(r.DisposeMessage)
	return nil
}
