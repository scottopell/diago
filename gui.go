package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
	"github.com/dustin/go-humanize"

	"github.com/remeh/diago/pprof"
	"github.com/remeh/diago/profile"
	prof "github.com/remeh/diago/profile"
)

type GUI struct {
	// data
	pprofProfile *pprof.Profile
	profile      *prof.Profile
	tree         *prof.FunctionsTree

	// ui options
	mode                prof.SampleMode
	searchField         string
	aggregateByFunction bool
}

func NewGUI(profile *pprof.Profile) *GUI {
	// init the base GUI object and load the profile
	// ----------------------

	g := &GUI{
		pprofProfile:        profile,
		aggregateByFunction: true,
	}
	g.reloadProfile()

	// depending on the profile opened, switch the either the ModeCpu
	// or the ModeHeapAlloc.
	// ----------------------

	switch prof.ReadProfileType(profile) {
	case "space":
		g.mode = prof.ModeHeapAlloc
	case "cpu":
		g.mode = prof.ModeCpu
	default:
		g.mode = prof.ModeDefault
	}

	return g
}

func (g *GUI) OpenWindow() {
	wnd := giu.NewMasterWindow("Diago", 800, 600, 0)
	wnd.Run(g.windowLoop)
}

func (g *GUI) onAggregationClick() {
	g.reloadProfile()
}

func (g *GUI) onAllocated() {
	g.mode = prof.ModeHeapAlloc
	g.reloadProfile()
}

func (g *GUI) onInuse() {
	g.mode = prof.ModeHeapInuse
	g.reloadProfile()
}

func (g *GUI) onSearch() {
	g.tree = g.profile.BuildTree(config.File, g.aggregateByFunction, g.searchField)
}

func (g *GUI) reloadProfile() {
	// read the pprof profile
	// ----------------------

	profile, err := prof.NewProfile(g.pprofProfile, g.mode)
	if err != nil {
		fmt.Println("err:", err)
		os.Exit(-1)
	}
	g.profile = profile

	// rebuild the displayed tree
	// ----------------------

	g.tree = g.profile.BuildTree(config.File, g.aggregateByFunction, g.searchField)
}

func (g *GUI) windowLoop() {
	giu.SingleWindow().Layout(
		g.toolbox(),
		g.treeFromFunctionsTree(g.tree),
	)
}

func (g *GUI) toolbox() *giu.RowWidget {
	size := giu.Context.GetPlatform().DisplaySize()

	widgets := make([]giu.Widget, 0)

	// search bar
	// ----------------------

	filterText := giu.InputText(&g.searchField).Flags(imgui.InputTextFlagsCallbackAlways).Label("Filter...").OnChange(g.onSearch).Size(size[0] / 4)

	widgets = append(widgets, filterText)

	// aggregate per func option
	// ----------------------
	widgets = append(widgets,
		giu.Checkbox("aggregate by functions", &g.aggregateByFunction).OnChange(g.onAggregationClick))
	widgets = append(widgets,
		giu.Tooltip("By default, Diago aggregates by functions, uncheck to have the information up to the lines of code"))

	// in heap mode, offer the two modes
	// ----------------------
	if g.mode == prof.ModeHeapAlloc || g.mode == prof.ModeHeapInuse {
		widgets = append(widgets,
			giu.RadioButton("allocated", g.mode == prof.ModeHeapAlloc).OnChange(g.onAllocated))
		widgets = append(widgets,
			giu.RadioButton("inuse", g.mode == prof.ModeHeapInuse).OnChange(g.onInuse))
	}

	return giu.Row(
		widgets...,
	)
}

func (g *GUI) treeFromFunctionsTree(tree *prof.FunctionsTree) giu.Layout {
	// generate the header
	// ----------------------

	var text string
	switch g.mode {
	case prof.ModeCpu:
		text = fmt.Sprintf("%s - total sampling duration: %s - total capture duration %s", tree.Name, time.Duration(g.profile.TotalSampling).String(), g.profile.CaptureDuration.String())
	case profile.ModeHeapAlloc:
		text = fmt.Sprintf("%s - total allocated memory: %s", tree.Name, humanize.IBytes(g.profile.TotalSampling))
	case profile.ModeHeapInuse:
		text = fmt.Sprintf("%s - total in-use memory: %s", tree.Name, humanize.IBytes(g.profile.TotalSampling))
	}

	// start generating the tree
	// ----------------------

	return giu.Layout{
		giu.Row(
			giu.TreeNode(text).Flags(giu.TreeNodeFlagsNone | giu.TreeNodeFlagsFramed | giu.TreeNodeFlagsDefaultOpen).Layout(g.treeNodeFromFunctionsTreeNode(tree.Root)),
		),
	}
}

func (g *GUI) treeNodeFromFunctionsTreeNode(node *prof.TreeNode) giu.Layout {
	if node == nil {
		return nil
	}
	rv := giu.Layout{}
	for _, child := range node.Children {
		if !child.Visible {
			continue
		}

		flags := giu.TreeNodeFlagsSpanAvailWidth
		if child.IsLeaf() {
			flags |= giu.TreeNodeFlagsLeaf
		}

		// generate the displayed texts
		// ----------------------
		_, _, tooltip, lineText := g.texts(child)

		// append the line to the tree
		// ----------------------

		rv = append(rv, giu.Row(
			giu.ProgressBar(float32(child.Percent)/100).Size(90, 0).Overlayf("%.3f%%", child.Percent),
			giu.Tooltip(tooltip),
			giu.TreeNode(lineText).Flags(flags).Layout(g.treeNodeFromFunctionsTreeNode(child)),
		),
		)
	}

	return rv
}

func (g *GUI) texts(node *prof.TreeNode) (value string, self string, tooltip string, lineText string) {
	if g.profile.Type == "cpu" {
		value = time.Duration(node.Value).String()
		self = time.Duration(node.Self).String()
		tooltip = fmt.Sprintf("%s of %s\nself: %s", value, time.Duration(g.profile.TotalSampling).String(), self)
	} else {
		value = humanize.IBytes(uint64(node.Value))
		self = humanize.IBytes(uint64(node.Self))
		tooltip = fmt.Sprintf("%s of %s\nself: %s", value, humanize.IBytes(g.profile.TotalSampling), self)
	}
	lineText = fmt.Sprintf("%s %s:%d - %s - self: %s", node.Function.Name, path.Base(node.Function.File), node.Function.LineNumber, value, self)
	if g.aggregateByFunction {
		lineText = fmt.Sprintf("%s %s - %s - self: %s", node.Function.Name, path.Base(node.Function.File), value, self)
	}
	return value, self, tooltip, lineText
}
