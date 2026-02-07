package theme

import "github.com/diamondburned/gotk4/pkg/gdk/v4"

// FontConfig holds all font size settings for the application
type FontConfig struct {
	HeaderSize   float64
	FilenameSize float64
	SizeTextSize float64
}

// SpacingConfig calculates dynamic spacing based on font sizes
type SpacingConfig struct {
	Fonts FontConfig
}

// NewTheme creates a new theme with default font sizes
func NewTheme() *SpacingConfig {
	return &SpacingConfig{
		Fonts: FontConfig{
			HeaderSize:   10,
			FilenameSize: 14,
			SizeTextSize: 11,
		},
	}
}

// GetFilenameYOffset returns the Y offset for filename text based on row height and font size
func (t *SpacingConfig) GetFilenameYOffset(rowHeight int) float64 {
	// Center text vertically with some padding
	return float64(rowHeight)/2 + t.Fonts.FilenameSize/3
}

// GetSizeTextYOffset returns the Y offset for size text based on row height and font size
func (t *SpacingConfig) GetSizeTextYOffset(rowHeight int) float64 {
	return float64(rowHeight)/2 + t.Fonts.SizeTextSize/3
}

// GetRowHeight returns appropriate row height based on font sizes and icon size
func (t *SpacingConfig) GetRowHeight(iconSize int) int {
	// Minimum row height based on icon size + padding
	minHeight := iconSize + 4
	// Or based on largest font + padding
	fontHeight := int(t.Fonts.FilenameSize) + 4
	if fontHeight > minHeight {
		minHeight = fontHeight
	}
	return minHeight
}

// Color theme
type ColorTheme struct {
	BackgroundColor       gdk.RGBA
	TextColor             gdk.RGBA
	SelectedBgColor       gdk.RGBA
	SelectedTextColor     gdk.RGBA
	HeaderBackgroundColor gdk.RGBA
	HeaderTextColor       gdk.RGBA
	CopyCutBgColor        gdk.RGBA
	HoverBgColor          gdk.RGBA
}

// NewColorTheme creates a new color theme with default dark colors
func NewColorTheme() *ColorTheme {
	return &ColorTheme{
		BackgroundColor:       gdk.NewRGBA(45.0/255, 45.0/255, 45.0/255, 1),
		TextColor:             gdk.NewRGBA(245.0/255, 245.0/255, 245.0/255, 1),
		SelectedBgColor:       gdk.NewRGBA(64.0/255, 64.0/255, 64.0/255, 1),
		SelectedTextColor:     gdk.NewRGBA(245.0/255, 245.0/255, 245.0/255, 1),
		HeaderBackgroundColor: gdk.NewRGBA(36.0/255, 36.0/255, 36.0/255, 1),
		HeaderTextColor:       gdk.NewRGBA(245.0/255, 245.0/255, 245.0/255, 1),
		CopyCutBgColor:        gdk.NewRGBA(50.0/255, 70.0/255, 90.0/255, 1),
		HoverBgColor:          gdk.NewRGBA(55.0/255, 55.0/255, 55.0/255, 1),
	}
}
