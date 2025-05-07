package services

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
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
	tempDir, err := os.MkdirTemp("", "pdf-generation")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Construct full template path
	templatePath := filepath.Join(g.templateDir, templateName)

	// Load CSS if provided
	var cssContent string
	if cssName != "" {
		cssPath := filepath.Join(g.cssDir, cssName)
		cssBytes, err := os.ReadFile(cssPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CSS file %s: %v", cssPath, err)
		}
		cssContent = string(cssBytes)
	}

	// Load the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %v", templatePath, err)
	}

	// Add CSS to the data if we have a template that supports it
	if data != nil {
		// Try to set CSS field if the data structure has it
		if dataMap, ok := data.(map[string]interface{}); ok {
			dataMap["CSS"] = cssContent
		}
	}

	// Create a temporary HTML file
	htmlFilePath := filepath.Join(tempDir, "output.html")
	htmlFile, err := os.Create(htmlFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create html file: %v", err)
	}

	// Execute the template
	err = tmpl.Execute(htmlFile, data)
	htmlFile.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}

	// Create PDF file path
	pdfFilePath := filepath.Join(tempDir, "output.pdf")

	// Execute wkhtmltopdf
	wkhtmltopdfArgs := []string{
		"--enable-local-file-access", // Allow access to local files (important for wkhtmltopdf)
		"--page-size", "A4",          // Set page size
		"--margin-top", "10mm", // Set margins
		"--margin-right", "10mm",
		"--margin-bottom", "10mm",
		"--margin-left", "10mm",
		"--print-media-type", // Use print media CSS
		htmlFilePath,         // Input HTML file
		pdfFilePath,          // Output PDF file
	}

	cmd := exec.Command(g.wkhtmltopdfPath, wkhtmltopdfArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("wkhtmltopdf failed: %v\nOutput: %s", err, string(output))
	}

	// Read the generated PDF
	pdfContent, err := os.ReadFile(pdfFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated PDF: %v", err)
	}

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
