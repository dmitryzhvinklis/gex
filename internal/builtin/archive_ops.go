package builtin

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Tar creates and extracts tar archives
func Tar(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("tar: missing operation")
	}

	var create, extract, list bool
	var verbose bool
	var gzipCompress bool
	var archive string
	var files []string

	// Parse arguments
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'c':
					create = true
				case 'x':
					extract = true
				case 't':
					list = true
				case 'v':
					verbose = true
				case 'z':
					gzipCompress = true
				case 'f':
					if i+1 < len(args) {
						archive = args[i+1]
					}
				}
			}
		} else if arg != archive {
			files = append(files, arg)
		}
	}

	if archive == "" {
		return fmt.Errorf("tar: missing archive file")
	}

	if create {
		return tarCreate(archive, files, verbose, gzipCompress)
	} else if extract {
		return tarExtract(archive, verbose, gzipCompress)
	} else if list {
		return tarList(archive, verbose, gzipCompress)
	}

	return fmt.Errorf("tar: no operation specified")
}

// tarCreate creates a tar archive
func tarCreate(archiveName string, files []string, verbose, gzipCompress bool) error {
	// Create archive file
	archiveFile, err := os.Create(archiveName)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	var writer io.Writer = archiveFile

	// Add gzip compression if requested
	var gzipWriter *gzip.Writer
	if gzipCompress {
		gzipWriter = gzip.NewWriter(archiveFile)
		writer = gzipWriter
		defer gzipWriter.Close()
	}

	// Create tar writer
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	// Add files to archive
	for _, file := range files {
		if err := addFileToTar(tarWriter, file, verbose); err != nil {
			return err
		}
	}

	return nil
}

// addFileToTar adds a file to tar archive
func addFileToTar(tarWriter *tar.Writer, filename string, verbose bool) error {
	return filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = path

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if verbose {
			fmt.Println(path)
		}

		// Write file content if it's a regular file
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			return err
		}

		return nil
	})
}

// tarExtract extracts a tar archive
func tarExtract(archiveName string, verbose, gzipCompress bool) error {
	// Open archive file
	archiveFile, err := os.Open(archiveName)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	var reader io.Reader = archiveFile

	// Handle gzip decompression
	if gzipCompress {
		gzipReader, err := gzip.NewReader(archiveFile)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Create tar reader
	tarReader := tar.NewReader(reader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if verbose {
			fmt.Println(header.Name)
		}

		// Create file/directory
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(header.Name, header.FileInfo().Mode()); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(header.Name), 0755); err != nil {
				return err
			}

			// Create file
			file, err := os.Create(header.Name)
			if err != nil {
				return err
			}

			// Copy content
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}
			file.Close()

			// Set permissions
			if err := os.Chmod(header.Name, header.FileInfo().Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}

// tarList lists contents of tar archive
func tarList(archiveName string, verbose, gzipCompress bool) error {
	// Open archive file
	archiveFile, err := os.Open(archiveName)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	var reader io.Reader = archiveFile

	// Handle gzip decompression
	if gzipCompress {
		gzipReader, err := gzip.NewReader(archiveFile)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Create tar reader
	tarReader := tar.NewReader(reader)

	// List files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if verbose {
			fmt.Printf("%s %10d %s %s\n",
				header.FileInfo().Mode(),
				header.Size,
				header.ModTime.Format("2006-01-02 15:04"),
				header.Name)
		} else {
			fmt.Println(header.Name)
		}
	}

	return nil
}

// Gzip compresses files using gzip
func Gzip(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("gzip: missing file")
	}

	var decompress bool
	var keep bool
	var files []string

	// Parse arguments
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'd':
					decompress = true
				case 'k':
					keep = true
				}
			}
		} else {
			files = append(files, arg)
		}
	}

	for _, file := range files {
		if decompress {
			if err := gunzipFile(file, keep); err != nil {
				fmt.Printf("gzip: %v\n", err)
			}
		} else {
			if err := gzipFile(file, keep); err != nil {
				fmt.Printf("gzip: %v\n", err)
			}
		}
	}

	return nil
}

// gzipFile compresses a file
func gzipFile(filename string, keep bool) error {
	// Open input file
	inputFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	// Create output file
	outputFile, err := os.Create(filename + ".gz")
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	// Copy data
	if _, err := io.Copy(gzipWriter, inputFile); err != nil {
		return err
	}

	// Remove original file if not keeping
	if !keep {
		return os.Remove(filename)
	}

	return nil
}

// gunzipFile decompresses a gzip file
func gunzipFile(filename string, keep bool) error {
	// Open input file
	inputFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(inputFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// Determine output filename
	outputName := filename
	if strings.HasSuffix(outputName, ".gz") {
		outputName = outputName[:len(outputName)-3]
	} else {
		outputName = outputName + ".out"
	}

	// Create output file
	outputFile, err := os.Create(outputName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Copy data
	if _, err := io.Copy(outputFile, gzipReader); err != nil {
		return err
	}

	// Remove original file if not keeping
	if !keep {
		return os.Remove(filename)
	}

	return nil
}

// Zip creates and extracts zip archives
func Zip(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("zip: missing arguments")
	}

	var extract bool
	var verbose bool
	var archive string
	var files []string

	// Parse arguments
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'x':
					extract = true
				case 'v':
					verbose = true
				}
			}
		} else {
			if archive == "" {
				archive = arg
			} else {
				files = append(files, args[i:]...)
				break
			}
		}
	}

	if archive == "" {
		return fmt.Errorf("zip: missing archive name")
	}

	if extract {
		return unzipArchive(archive, verbose)
	} else {
		return createZipArchive(archive, files, verbose)
	}
}

// createZipArchive creates a zip archive
func createZipArchive(archiveName string, files []string, verbose bool) error {
	// Create archive file
	archiveFile, err := os.Create(archiveName)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	// Add files to archive
	for _, file := range files {
		if err := addFileToZip(zipWriter, file, verbose); err != nil {
			return err
		}
	}

	return nil
}

// addFileToZip adds a file to zip archive
func addFileToZip(zipWriter *zip.Writer, filename string, verbose bool) error {
	return filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Open source file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Create zip file entry
		writer, err := zipWriter.Create(path)
		if err != nil {
			return err
		}

		// Copy file content
		if _, err := io.Copy(writer, file); err != nil {
			return err
		}

		if verbose {
			fmt.Printf("adding: %s\n", path)
		}

		return nil
	})
}

// unzipArchive extracts a zip archive
func unzipArchive(archiveName string, verbose bool) error {
	// Open zip file
	reader, err := zip.OpenReader(archiveName)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Extract files
	for _, file := range reader.File {
		if verbose {
			fmt.Printf("extracting: %s\n", file.Name)
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(file.Name), 0755); err != nil {
			return err
		}

		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Open file in zip
		rc, err := file.Open()
		if err != nil {
			return err
		}

		// Create destination file
		outFile, err := os.Create(file.Name)
		if err != nil {
			rc.Close()
			return err
		}

		// Copy content
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
