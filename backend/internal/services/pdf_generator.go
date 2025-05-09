package services

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// PDFGenerator handles the generation of PDF documents
type PDFGenerator struct {
	templateDir     string
	cssDir          string
	wkhtmltopdfPath string
}

// NewPDFGenerator creates a new PDF generator service
func NewPDFGenerator(templateDir, cssDir, wkhtmltopdfPath string) *PDFGenerator {
	return &PDFGenerator{
		templateDir:     templateDir,
		cssDir:          cssDir,
		wkhtmltopdfPath: wkhtmltopdfPath,
	}
}

// GenerateFromTemplate generates a PDF from a template with given data
func (g *PDFGenerator) GenerateFromTemplate(templateName string, cssName string, data interface{}) ([]byte, error) {
	// Create a temporary directory for our files
	log.Printf("Starting PDF generation for template: %s", templateName)
	tempDir, err := os.MkdirTemp("", "pdf-generation")
	if err != nil {
		log.Printf("ERROR: Failed to create temp directory: %v", err)
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	log.Printf("Created temp directory: %s", tempDir)

	// Construct full template path
	templatePath := filepath.Join(g.templateDir, templateName)
	log.Printf("Template path: %s", templatePath)

	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		log.Printf("ERROR: Template file does not exist: %s", templatePath)
		return nil, fmt.Errorf("template file does not exist: %s", templatePath)
	}

	// Load CSS if provided
	var cssContent string
	if cssName != "" {
		cssPath := filepath.Join(g.cssDir, cssName)
		log.Printf("CSS path: %s", cssPath)

		// Check if CSS file exists
		if _, err := os.Stat(cssPath); os.IsNotExist(err) {
			log.Printf("ERROR: CSS file does not exist: %s", cssPath)
			return nil, fmt.Errorf("CSS file does not exist: %s", cssPath)
		}

		cssBytes, err := os.ReadFile(cssPath)
		if err != nil {
			log.Printf("ERROR: Failed to read CSS file: %v", err)
			return nil, fmt.Errorf("failed to read CSS file %s: %v", cssPath, err)
		}
		cssContent = string(cssBytes)
		log.Printf("CSS file loaded, length: %d bytes", len(cssContent))
	}

	// Load the template
	log.Printf("Parsing template file")
	// Create a new template with functions
	tmpl := template.New(filepath.Base(templatePath)).Funcs(template.FuncMap{
		"formatMoney": func(amount float64) string {
			// Format with two decimal places
			formattedAmount := fmt.Sprintf("%.2f", amount)

			// Split into integer and decimal parts
			parts := strings.Split(formattedAmount, ".")
			integerPart := parts[0]
			decimalPart := parts[1]

			// Add thousand separators to integer part
			for i := len(integerPart) - 3; i > 0; i -= 3 {
				integerPart = integerPart[:i] + "," + integerPart[i:]
			}

			return integerPart + "." + decimalPart
		},
		"calculateDiscountPercent": func(quantity interface{}, unitPrice, discount interface{}) string {
			// Output debug information
			log.Printf("DEBUG: calculateDiscountPercent input - quantity: %v, unitPrice: %v, discount: %v", quantity, unitPrice, discount)

			// Convert parameters to float64 safely
			q := 0.0
			up := 0.0
			d := 0.0

			// Convert quantity
			switch v := quantity.(type) {
			case int:
				q = float64(v)
			case float64:
				q = v
			case int64:
				q = float64(v)
			default:
				log.Printf("DEBUG: Unknown quantity type: %T", quantity)
			}

			// Convert unit price
			switch v := unitPrice.(type) {
			case float64:
				up = v
			case int:
				up = float64(v)
			case string:
				f, err := strconv.ParseFloat(v, 64)
				if err == nil {
					up = f
				}
			default:
				log.Printf("DEBUG: Unknown unitPrice type: %T", unitPrice)
			}

			// Convert discount
			switch v := discount.(type) {
			case float64:
				d = v
			case int:
				d = float64(v)
			case string:
				f, err := strconv.ParseFloat(v, 64)
				if err == nil {
					d = f
				}
			default:
				log.Printf("DEBUG: Unknown discount type: %T", discount)
			}

			log.Printf("DEBUG: After conversion - q: %v, up: %v, d: %v", q, up, d)

			// Check for zero values
			if d <= 0 {
				return "-"
			}

			// Calculate line total
			lineBeforeDiscount := q * up
			if lineBeforeDiscount <= 0 {
				return "-"
			}

			// FIXED LOGIC: If discount seems to be a percentage already (0-100 range)
			// and it's much smaller than the line total, treat it as a direct percentage
			if d > 0 && d <= 100 && d < (lineBeforeDiscount*0.1) {
				// Treat the value as a direct percentage (e.g., 50 means 50%)
				log.Printf("DEBUG: Treating discount as a direct percentage: %v%%", d)
				return fmt.Sprintf("%.1f%%", d)
			}

			// Otherwise calculate as monetary discount
			percent := (d / lineBeforeDiscount) * 100
			log.Printf("DEBUG: Calculated as monetary discount, percent: %v", percent)

			// Format based on size
			if percent < 0.1 {
				return fmt.Sprintf("%.4f%%", percent)
			} else {
				return fmt.Sprintf("%.1f%%", percent)
			}
		},
	})

	// Parse the template file
	tmpl, err = tmpl.ParseFiles(templatePath)
	if err != nil {
		log.Printf("ERROR: Failed to parse template: %v", err)
		return nil, fmt.Errorf("failed to parse template %s: %v", templatePath, err)
	}

	// Add CSS to the data if we have a template that supports it
	if data != nil {
		// Try to set CSS field if the data structure has it
		if dataMap, ok := data.(map[string]interface{}); ok {
			dataMap["CSS"] = cssContent
			log.Printf("Added CSS to template data: %d bytes", len(cssContent))
		} else {
			log.Printf("WARNING: Cannot add CSS to template data - data is not a map[string]interface{}")
		}
	} else {
		// If data is nil, create a new map with just the CSS
		data = map[string]interface{}{
			"CSS": cssContent,
		}
		log.Printf("Created new data map with CSS")
	}

	// Create a temporary HTML file
	htmlFilePath := filepath.Join(tempDir, "output.html")
	log.Printf("Creating HTML file: %s", htmlFilePath)
	htmlFile, err := os.Create(htmlFilePath)
	if err != nil {
		log.Printf("ERROR: Failed to create HTML file: %v", err)
		return nil, fmt.Errorf("failed to create html file: %v", err)
	}

	// Execute the template
	log.Printf("Executing template with data")
	err = tmpl.Execute(htmlFile, data)
	htmlFile.Close()
	if err != nil {
		log.Printf("ERROR: Failed to execute template: %v", err)
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}
	log.Printf("Template executed successfully")

	// Create PDF file path
	pdfFilePath := filepath.Join(tempDir, "output.pdf")
	log.Printf("PDF output path: %s", pdfFilePath)

	// Execute wkhtmltopdf
	wkhtmltopdfArgs := []string{
		"--quiet",                    // Reduce output noise
		"--enable-local-file-access", // Allow access to local files (important for wkhtmltopdf)
		htmlFilePath,                 // Input HTML file
		pdfFilePath,                  // Output PDF file
	}

	log.Printf("Executing wkhtmltopdf: %s %s", g.wkhtmltopdfPath, strings.Join(wkhtmltopdfArgs, " "))
	cmd := exec.Command(g.wkhtmltopdfPath, wkhtmltopdfArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: wkhtmltopdf failed: %v\nCommand output: %s", err, string(output))
		return nil, fmt.Errorf("wkhtmltopdf failed: %v\nOutput: %s", err, string(output))
	}
	log.Printf("wkhtmltopdf executed successfully")

	// Read the generated PDF
	log.Printf("Reading generated PDF file")
	pdfContent, err := os.ReadFile(pdfFilePath)
	if err != nil {
		log.Printf("ERROR: Failed to read generated PDF: %v", err)
		return nil, fmt.Errorf("failed to read generated PDF: %v", err)
	}
	log.Printf("PDF file read successfully, size: %d bytes", len(pdfContent))

	return pdfContent, nil
}

// Detect attempts to find the wkhtmltopdf binary in standard locations
func DetectWkhtmltopdfPath() string {
	// Common locations for wkhtmltopdf
	locations := []string{
		"wkhtmltopdf",                // Available in PATH
		"/usr/bin/wkhtmltopdf",       // Linux
		"/usr/local/bin/wkhtmltopdf", // Linux/macOS
		"C:\\Program Files\\wkhtmltopdf\\bin\\wkhtmltopdf.exe",       // Windows
		"C:\\Program Files (x86)\\wkhtmltopdf\\bin\\wkhtmltopdf.exe", // Windows (x86)
	}

	for _, loc := range locations {
		cmd := exec.Command(loc, "--version")
		err := cmd.Run()
		if err == nil {
			return loc
		}
	}

	// Default to PATH if we can't find it
	return "wkhtmltopdf"
}

// EnsureTemplateDirectories ensures that the template directories exist
func EnsureTemplateDirectories(baseDir, cssDir, templateDir string) error {
	dirs := []string{
		baseDir,
		filepath.Join(baseDir, cssDir),
	}

	// Add template directories (they could be nested)
	if templateDir != "" {
		templateDirParts := strings.Split(templateDir, string(os.PathSeparator))
		currentPath := baseDir
		for _, part := range templateDirParts {
			currentPath = filepath.Join(currentPath, part)
			dirs = append(dirs, currentPath)
		}
	}

	// Create directories if they don't exist
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	return nil
}
