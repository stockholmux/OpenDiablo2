package d2maprenderer

import (
	"math"

	"github.com/OpenDiablo2/OpenDiablo2/d2common"
)

type worldTrans struct {
	x float64
	y float64
}

const (
	center = 0
	left   = 1
	right  = 2
)

// Viewport is used for converting vectors between screen (pixel), orthogonal (camera) and world (isometric) space.
// TODO: Has a coordinate (issue #456)
type Viewport struct {
	defaultScreenRect d2common.Rectangle
	screenRect        d2common.Rectangle
	transStack        []worldTrans
	transCurrent      worldTrans
	camera            *Camera
	align             int
}

// NewViewport creates a new Viewport with the given parameters and returns a pointer to it.
func NewViewport(x, y, width, height int) *Viewport {
	return &Viewport{
		screenRect: d2common.Rectangle{
			Left:   x,
			Top:    y,
			Width:  width,
			Height: height,
		},
		defaultScreenRect: d2common.Rectangle{
			Left:   x,
			Top:    y,
			Width:  width,
			Height: height,
		},
	}
}

// SetCamera sets the current camera to the given value.
func (v *Viewport) SetCamera(camera *Camera) {
	v.camera = camera
}

// WorldToScreen returns the screen space for the given world coordinates as two integers.
func (v *Viewport) WorldToScreen(x, y float64) (int, int) {
	return v.OrthoToScreen(v.WorldToOrtho(x, y))
}

// WorldToScreenF returns the screen space for the given world coordinates as two float64s.
func (v *Viewport) WorldToScreenF(x, y float64) (float64, float64) {
	return v.OrthoToScreenF(v.WorldToOrtho(x, y))
}

// ScreenToWorld returns the world position for the given screen coordinates.
func (v *Viewport) ScreenToWorld(x, y int) (float64, float64) {
	return v.OrthoToWorld(v.ScreenToOrtho(x, y))
}

// OrthoToWorld returns the world position for the given orthogonal coordinates.
func (v *Viewport) OrthoToWorld(x, y float64) (float64, float64) {
	worldX := (x/80 + y/40) / 2
	worldY := (y/40 - x/80) / 2

	return worldX, worldY
}

// WorldToOrtho returns the orthogonal position for the given world coordinates.
func (v *Viewport) WorldToOrtho(x, y float64) (float64, float64) {
	orthoX := (x - y) * 80
	orthoY := (x + y) * 40

	return orthoX, orthoY
}

// ScreenToOrtho returns the orthogonal position for the given screen coordinates.
func (v *Viewport) ScreenToOrtho(x, y int) (float64, float64) {
	camX, camY := v.getCameraOffset()
	screenX := float64(x) + camX - float64(v.screenRect.Left)
	screenY := float64(y) + camY - float64(v.screenRect.Top)

	return screenX, screenY
}

// OrthoToScreen returns the screen position for the given orthogonal coordinates as two ints.
func (v *Viewport) OrthoToScreen(x, y float64) (int, int) {
	camOrthoX, camOrthoY := v.getCameraOffset()
	orthoX := int(math.Floor(x - camOrthoX + float64(v.screenRect.Left)))
	orthoY := int(math.Floor(y - camOrthoY + float64(v.screenRect.Top)))

	return orthoX, orthoY
}

// OrthoToScreenF returns the screen position for the given orthogonal coordinates as two float64s.
func (v *Viewport) OrthoToScreenF(x, y float64) (float64, float64) {
	camOrthoX, camOrthoY := v.getCameraOffset()
	orthoX := x - camOrthoX + float64(v.screenRect.Left)
	orthoY := y - camOrthoY + float64(v.screenRect.Top)

	return orthoX, orthoY
}

// IsTileVisible returns false if no part of the tile is within the game screen.
func (v *Viewport) IsTileVisible(x, y float64) bool {
	orthoX1, orthoY1 := v.WorldToOrtho(x-3, y)
	orthoX2, orthoY2 := v.WorldToOrtho(x+3, y)

	return v.IsOrthoRectVisible(orthoX1, orthoY1, orthoX2, orthoY2)
}

// IsTileRectVisible returns false if none of the tiles rects are within the game screen.
func (v *Viewport) IsTileRectVisible(rect d2common.Rectangle) bool {
	left := float64((rect.Left - rect.Bottom()) * 80)
	top := float64((rect.Left + rect.Top) * 40)
	right := float64((rect.Right() - rect.Top) * 80)
	bottom := float64((rect.Right() + rect.Bottom()) * 40)

	return v.IsOrthoRectVisible(left, top, right, bottom)
}

// IsOrthoRectVisible returns false if the given orthogonal position is outside the game screen.
func (v *Viewport) IsOrthoRectVisible(x1, y1, x2, y2 float64) bool {
	screenX1, screenY1 := v.OrthoToScreen(x1, y1)
	screenX2, screenY2 := v.OrthoToScreen(x2, y2)

	return !(screenX1 >= v.defaultScreenRect.Width || screenX2 < 0 || screenY1 >= v.defaultScreenRect.Height || screenY2 < 0)
}

// GetTranslationOrtho returns the viewport's current orthogonal space translation.
func (v *Viewport) GetTranslationOrtho() (float64, float64) {
	return v.transCurrent.x, v.transCurrent.y
}

// GetTranslationScreen returns the viewport's current screen space translation.
func (v *Viewport) GetTranslationScreen() (int, int) {
	return v.OrthoToScreen(v.transCurrent.x, v.transCurrent.y)
}

// PushTranslationOrtho adds a new orthogonal translation to the stack.
func (v *Viewport) PushTranslationOrtho(x, y float64) *Viewport {
	v.transStack = append(v.transStack, v.transCurrent)
	v.transCurrent.x += x
	v.transCurrent.y += y

	return v
}

// PushTranslationWorld adds a new world translation to the stack, converting it to orthogonal space.
func (v *Viewport) PushTranslationWorld(x, y float64) {
	v.PushTranslationOrtho(v.WorldToOrtho(x, y))
}

// PushTranslationScreen adds a new screen translation to the stack, converting it to orthogonal space.
func (v *Viewport) PushTranslationScreen(x, y int) {
	v.PushTranslationOrtho(v.ScreenToOrtho(x, y))
}

// PopTranslation pops a translation from the stack.
func (v *Viewport) PopTranslation() {
	count := len(v.transStack)
	if count == 0 {
		panic("empty stack")
	}

	v.transCurrent = v.transStack[count-1]
	v.transStack = v.transStack[:count-1]
}

func (v *Viewport) getCameraOffset() (float64, float64) {
	var camX, camY float64
	if v.camera != nil {
		camX, camY = v.camera.GetPosition()
	}

	camX -= float64(v.screenRect.Width / 2)
	camY -= float64(v.screenRect.Height / 2)

	return camX, camY
}

func (v *Viewport) toLeft() {
	if v.align == left {
		return
	}

	v.screenRect.Width = v.defaultScreenRect.Width / 2
	v.screenRect.Left = v.defaultScreenRect.Left + v.defaultScreenRect.Width/2
	v.align = left
}

func (v *Viewport) toRight() {
	if v.align == right {
		return
	}

	v.screenRect.Width = v.defaultScreenRect.Width / 2
	v.align = right
}

func (v *Viewport) resetAlign() {
	if v.align == center {
		return
	}

	v.screenRect.Width = v.defaultScreenRect.Width
	v.screenRect.Left = v.defaultScreenRect.Left
	v.align = center
}
