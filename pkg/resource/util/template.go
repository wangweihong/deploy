package util

import (
	"bytes"

	"ufleet-deploy/pkg/log"

	"k8s.io/apimachinery/pkg/runtime"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/printers"
)

func GetYamlTemplateFromObject(origin runtime.Object) (*string, error) {

	f := cmdutil.NewFactory(nil)
	log.DebugPrint(origin.GetObjectKind().GroupVersionKind().String())
	//	log.DebugPrint(origin)

	//一定要编码,不然会出现字段首字母全大写的情况
	//1.7客户端返回的TypeMeta为空,无法进行Encoder
	encoder := f.JSONEncoder()
	data, err := runtime.Encode(encoder, origin)
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
