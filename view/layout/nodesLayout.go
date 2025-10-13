package layout

import "fyne.io/fyne/v2"

type NodeRecordsGrid struct{}

func (n *NodeRecordsGrid) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	for _, o := range objects {
		childSize := o.MinSize()
		w += childSize.Width
		h = childSize.Height
	}
	return fyne.NewSize(w, h)
}

func (n *NodeRecordsGrid) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := 0
	// ip/user/password/status/arch/kernel
	widths := []int{120, 100, 150, 80, 60, int(size.Width) - 490}
	for i, o := range objects {
		w := widths[i]
		o.Resize(fyne.NewSize(float32(w), size.Height))
		o.Move(fyne.NewPos(float32(x), 0))
		x += w
	}
}
