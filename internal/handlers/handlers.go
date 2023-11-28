package handlers

import (
	"ESP32-SPIFFS-burner/internal/request"
	"ESP32-SPIFFS-burner/internal/response"
	"ESP32-SPIFFS-burner/internal/utils"
	"ESP32-SPIFFS-burner/internal/validator"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Handlers struct {
	*utils.Utils
}

func NewHandlers(utils *utils.Utils) Handlers {
	return Handlers{Utils: utils}
}

func (h Handlers) Home(w http.ResponseWriter, r *http.Request) {
	var form struct {
		SourceDirectory string              `form:"SourceDirectory"`
		OutputDirectory string              `form:"OutputDirectory"`
		PageSize        string              `form:"PageSize"`
		BlockSize       string              `form:"BlockSize"`
		PartitionSize   string              `form:"PartitionSize"`
		ImageName       string              `form:"ImageName"`
		TargetChip      string              `form:"TargetChip"`
		SerialPort      string              `form:"SerialPort"`
		BaudRate        string              `form:"BaudRate"`
		FlashMode       string              `form:"FlashMode"`
		FlashFrequency  string              `form:"FlashFrequency"`
		Offset          string              `form:"Offset"`
		Output          template.HTML       `form:"Output"`
		ShowModal       bool                `form:"ShowModal"`
		Validator       validator.Validator `form:"-"`
	}

	switch r.Method {
	case http.MethodGet:

		err := response.Page(w, http.StatusOK, nil, "pages/home.html")
		if err != nil {
			h.ServerError(w, r, err)
		}

	case http.MethodPost:

		err := request.DecodePostForm(r, &form)
		if err != nil {
			h.BadRequest(w, r, err)
			return
		}

		form.Validator.CheckField(len(form.SourceDirectory) != 0, "SourceDirectory", "required")
		form.Validator.CheckField(len(form.OutputDirectory) != 0, "OutputDirectory", "required")
		form.Validator.CheckField(len(form.PageSize) != 0, "PageSize", "required")
		form.Validator.CheckField(len(form.BlockSize) != 0, "BlockSize", "required")
		form.Validator.CheckField(len(form.PartitionSize) != 0, "PartitionSize", "required")
		form.Validator.CheckField(len(form.ImageName) != 0, "ImageName", "required")
		form.Validator.CheckField(len(form.TargetChip) != 0, "TargetChip", "required")
		form.Validator.CheckField(len(form.SerialPort) != 0, "SerialPort", "required")
		form.Validator.CheckField(len(form.BaudRate) != 0, "BaudRate", "required")
		form.Validator.CheckField(len(form.FlashMode) != 0, "FlashMode", "required")
		form.Validator.CheckField(len(form.FlashFrequency) != 0, "FlashFrequency", "required")
		form.Validator.CheckField(len(form.Offset) != 0, "Offset", "required")

		data := h.newTemplateData(r)
		data["Form"] = form

		if form.Validator.HasErrors() {
			err := response.Page(w, http.StatusUnprocessableEntity, data, "pages/home.html")
			if err != nil {
				h.ServerError(w, r, err)
			}
			return
		}

		mkspiffsOutput, err := createSPIFFSImage(
			form.SourceDirectory,
			form.OutputDirectory,
			form.ImageName,
			form.PageSize,
			form.BlockSize,
			form.PartitionSize,
		)
		if err != nil {
			h.ServerError(w, r, err)
			return
		}

		esptoolOutput, err := burnSPIFFSImage(
			form.OutputDirectory,
			form.ImageName,
			form.TargetChip,
			form.SerialPort,
			form.BaudRate,
			form.FlashMode,
			form.FlashFrequency,
			form.Offset,
		)
		if err != nil {
			h.ServerError(w, r, err)
			return
		}

		mkspiffsHTMLContent := fmt.Sprintf("<strong>mkspiffs messages:</strong><br/>%s%s", mkspiffsOutput, "<br/>")
		esptoolHTMLContent := fmt.Sprintf("<strong>esptool messages:</strong><br/>%s%s", esptoolOutput, "<br/>")
		output := []string{mkspiffsHTMLContent, esptoolHTMLContent}
		outputJoin := strings.Join(output, " ")

		form.Output = template.HTML(outputJoin)
		form.ShowModal = true
		data["Form"] = form

		err = response.Page(w, http.StatusOK, data, "pages/home.html")
		if err != nil {
			h.ServerError(w, r, err)
		}

		return
	}
}

func (h Handlers) newTemplateData(_ *http.Request) map[string]any {
	data := make(map[string]any)
	return data
}

func createSPIFFSImage(
	sourceDirectory,
	outputDirectory,
	imageName,
	pageSize,
	blockSize,
	partitionSize string,
) (string, error) {
	cmdExec := "mkspiffs"

	path, err := lookForCommand(cmdExec)
	if err != nil {
		return "", err
	}

	args := fmt.Sprintf(
		"-c %s -p %s -b %s -s %s %s%s%s",
		sourceDirectory, pageSize, blockSize, partitionSize, outputDirectory, string(os.PathSeparator), imageName,
	)

	return execCommand(path, args)
}

func burnSPIFFSImage(
	outputDirectory,
	imageName,
	targetChip,
	serialPort,
	baudRate,
	flashMode,
	flashFrequency,
	partitionOffset string) (string, error) {

	cmdExec := "esptool"

	path, err := lookForCommand(cmdExec)
	if err != nil {
		return "", err
	}

	args := fmt.Sprintf(
		"--chip %s --port %s --baud %s write_flash -z --flash_mode %s --flash_freq %s --flash_size detect %s %s%s%s",
		targetChip, serialPort, baudRate, flashMode, fmt.Sprintf("%s%s", flashFrequency, "m"), partitionOffset, outputDirectory, string(os.PathSeparator), imageName,
	)

	return execCommand(path, args)
}

func lookForCommand(command string) (string, error) {
	path, err := exec.LookPath(command)
	if err != nil {
		fmt.Printf("Error: %s command not found in PATH\n", command)
		return "", err
	}

	return path, nil
}

func execCommand(path, args string) (string, error) {
	argList := strings.Fields(args)

	cmd := exec.Command(path, argList...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing %s: %v\n%s", path, err, out)
	}

	fmt.Printf("Command output:\n%s\n", string(out))

	return string(out), err
}
