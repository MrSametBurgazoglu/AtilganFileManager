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
			HeaderSize:   11,
			FilenameSize: 13,
			SizeTextSize: 10,
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
	AccentColor           gdk.RGBA
}

// NewColorTheme creates a new color theme with Aurora Dark palette
func NewColorTheme() *ColorTheme {
	return &ColorTheme{
		BackgroundColor:       gdk.NewRGBA(30.0/255, 30.0/255, 46.0/255, 1),    // #1e1e2e
		TextColor:             gdk.NewRGBA(205.0/255, 214.0/255, 244.0/255, 1), // #cdd6f4
		SelectedBgColor:       gdk.NewRGBA(138.0/255, 173.0/255, 244.0/255, 0.2),
		SelectedTextColor:     gdk.NewRGBA(137.0/255, 180.0/255, 250.0/255, 1), // #89b4fa
		HeaderBackgroundColor: gdk.NewRGBA(24.0/255, 24.0/255, 37.0/255, 1),    // #181825
		HeaderTextColor:       gdk.NewRGBA(186.0/255, 194.0/255, 222.0/255, 1), // #bac2de
		CopyCutBgColor:        gdk.NewRGBA(49.0/255, 50.0/255, 68.0/255, 1),    // #313244
		HoverBgColor:          gdk.NewRGBA(69.0/255, 71.0/255, 90.0/255, 0.1),
		AccentColor:           gdk.NewRGBA(137.0/255, 180.0/255, 250.0/255, 1),
	}
}
