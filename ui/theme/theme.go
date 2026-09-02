// Package theme provides color palettes, typography, and styling systems for Alap UI.
package theme

// ColorPalette defines the application color tokens.
type ColorPalette struct {
	Primary          string `json:"primary"`
	PrimaryContainer string `json:"primaryContainer"`
	Secondary        string `json:"secondary"`
	Background       string `json:"background"`
	Surface          string `json:"surface"`
	SurfaceVariant   string `json:"surfaceVariant"`
	TextPrimary      string `json:"textPrimary"`
	TextSecondary    string `json:"textSecondary"`
	Error            string `json:"error"`
	Success          string `json:"success"`
	Warning          string `json:"warning"`
	Border           string `json:"border"`
}

// Theme encapsulates light and dark color schemes and spacing tokens.
type Theme struct {
	IsDark       bool         `json:"isDark"`
	Colors       ColorPalette `json:"colors"`
	CornerRadius float64      `json:"cornerRadius"`
	FontFamily   string       `json:"fontFamily"`
}

// LightTheme returns the default Alap light theme.
func LightTheme() *Theme {
	return &Theme{
		IsDark: false,
		Colors: ColorPalette{
			Primary:          "#176BFF",
			PrimaryContainer: "#E6F0FF",
			Secondary:        "#5856D6",
			Background:       "#F5F5F7",
			Surface:          "#FFFFFF",
			SurfaceVariant:   "#EBEBEF",
			TextPrimary:      "#1C1C1E",
			TextSecondary:    "#8E8E93",
			Error:            "#FF3B30",
			Success:          "#34C759",
			Warning:          "#FF9500",
			Border:           "#D1D1D6",
		},
		CornerRadius: 10.0,
		FontFamily:   "system-ui, -apple-system, Roboto, sans-serif",
	}
}

// DarkTheme returns the default Alap dark theme.
func DarkTheme() *Theme {
	return &Theme{
		IsDark: true,
		Colors: ColorPalette{
			Primary:          "#388AF6",
			PrimaryContainer: "#0A2540",
			Secondary:        "#7D7AFF",
			Background:       "#000000",
			Surface:          "#1C1C1E",
			SurfaceVariant:   "#2C2C2E",
			TextPrimary:      "#FFFFFF",
			TextSecondary:    "#98989D",
			Error:            "#FF453A",
			Success:          "#32D74B",
			Warning:          "#FF9F0A",
			Border:           "#38383A",
		},
		CornerRadius: 10.0,
		FontFamily:   "system-ui, -apple-system, Roboto, sans-serif",
	}
}
