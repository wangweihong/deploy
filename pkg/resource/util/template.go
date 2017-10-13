package util

import (
	"bytes"
	"encoding/json"
	"io"

	"ufleet-deploy/pkg/log"

	ghyaml "github.com/ghodss/yaml"

	"k8s.io/apimachinery/pkg/util/yaml"

	"k8s.io/apimachinery/pkg/runtime"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/printers"
)

func GetObjectFromYamlTemplate(template []byte, obj interface{}) error {
	/*
		f := cmdutil.NewFactory(nil)
		decoder := f.Decoder(false)
		err := runtime.DecodeInto(decoder, template, obj)
		if err != nil {
			return err
		}
		return nil
	*/
	data, err := ghyaml.JSONToYAML(template)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &obj)

}

func GetYamlTemplateFromObject(origin runtime.Object) (*string, error) {

	json, err := json.Marshal(origin)
	if err != nil {
		return nil, err
	}
	data, err := ghyaml.JSONToYAML(json)
	if err != nil {
		return nil, err
	}
	str := string(data)
	return &str, err

	f := cmdutil.NewFactory(nil)
	log.DebugPrint(origin.GetObjectKind().GroupVersionKind().String())
	//	log.DebugPrint(origin)

	//一定要编码,不然会出现字段首字母全大写的情况
	//1.7客户端返回的TypeMeta为空,无法进行Encoder
	encoder := f.JSONEncoder()
	data, err = runtime.Encode(encoder, origin)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	//注意这里.
	obj := runtime.Unknown{Raw: data}

	buf := new(bytes.Buffer)
	//printer, _, _ := kubectl.GetPrinter("yaml", "", false)
	printer := printers.YAMLPrinter{}
	err = printer.PrintObj(&obj, buf)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	t := string(buf.Bytes())

	return &t, nil

}

const yamlSeparator = "\n---"
const separator = "---"

// splitYAMLDocument is a bufio.SplitFunc for splitting YAML streams into individual documents.
func splitYAMLDocument(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	sep := len([]byte(yamlSeparator))
	if i := bytes.Index(data, []byte(yamlSeparator)); i >= 0 {
		// We have a potential document terminator
		i += sep
		after := data[i:]
		if len(after) == 0 {
			// we can't read any more characters
			if atEOF {
				return len(data), data[:len(data)-sep], nil
			}
			return 0, nil, nil
		}
		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i-sep], nil
		}
		return 0, nil, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func ParseJsonOrYaml(data []byte) ([]runtime.RawExtension, error) {
	reader := bytes.NewReader(data)
	d := yaml.NewYAMLOrJSONDecoder(reader, 4096)

	exts := make([]runtime.RawExtension, 0)
	for {

		ext := runtime.RawExtension{}
		if err := d.Decode(&ext); err != nil {
			if err == io.EOF {
				return exts, nil
			}
			return nil, err
		}
		// TODO: This needs to be able to handle object in other encodings and schemas.
		ext.Raw = bytes.TrimSpace(ext.Raw)
		if len(ext.Raw) == 0 || bytes.Equal(ext.Raw, []byte("null")) {
			continue
		}
		exts = append(exts, ext)
	}

}
