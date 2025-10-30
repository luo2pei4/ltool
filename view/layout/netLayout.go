package layout

import "fyne.io/fyne/v2"

type NetRecordsGrid struct{}

type NIDAreaGrid struct{}

func (n *NetRecordsGrid) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	for _, o := range objects {
		childSize := o.MinSize()
		w += childSize.Width
		h = childSize.Height
	}
	return fyne.NewSize(w, h)
}

func (n *NetRecordsGrid) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := 0
	// name/ipv4/mac/type/state/nid
	widths := []int{100, 120, 150, 80, 60, int(size.Width) - 490}
	for i, o := range objects {
		w := widths[i]
		o.Resize(fyne.NewSize(float32(w), size.Height))
		o.Move(fyne.NewPos(float32(x), 0))
		x += w
	}
}

func (n *NIDAreaGrid) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	for _, o := range objects {
		childSize := o.MinSize()
		w += childSize.Width
		h = childSize.Height
	}
	return fyne.NewSize(w, h)
}

func (n *NIDAreaGrid) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := 0
	// ipv4/net type/idx
	widths := []int{130, 90, int(size.Width) - 220}
	for i, o := range objects {
		w := widths[i]
		o.Resize(fyne.NewSize(float32(w), size.Height))
		o.Move(fyne.NewPos(float32(x), 0))
		x += w
	}
}
