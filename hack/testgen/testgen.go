package main

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"golang.org/x/tools/txtar"
)

func main() {
	testCases := map[string]struct {
		Template string
		Payloads []any
	}{
		"describe": {
			Template: `
# describe {{ .Target }}
exec ttrpcurl --proto {{ .Proto }} describe {{ .Target }}
cmp stdout {{ .Proto }}.{{ .Target }}.out
! stderr .+
-- {{ .Proto }}.{{ .Target }}.out --
`,
			Payloads: sliceOfAny([]struct {
				Proto  string
				Target string
			}{
				{"product.proto", "product.ProductService"},
				{"product.proto", "product.ProductService.GetProduct"},
				{"product.proto", "product.ProductService.ListProducts"},
				{"product.proto", "product.ProductService.CreateProduct"},
				{"product.proto", "product.ProductService.UpdateProduct"},
				{"product.proto", "product.ProductService.DeleteProduct"},
			}),
		},
	}

	var caseBuf bytes.Buffer
	var result txtar.Archive
	for _, testcase := range testCases {
		for _, payload := range testcase.Payloads {

			tar := txtar.Parse([]byte(testcase.Template))

			commentTemp, err := template.New("comment").Parse(string(tar.Comment))
			if err != nil {
				log.Fatal(fmt.Errorf("parsing comment template: %w", err))
			}

			if err := commentTemp.Execute(&caseBuf, payload); err != nil {
				log.Fatal(fmt.Errorf("executing comment template: %w", err))
			}

			for _, file := range tar.Files {
				fileTemp, err := template.New("file").Parse(string(file.Name))
				if err != nil {
					log.Fatal(fmt.Errorf("parsing file template: %w", err))
				}

				var fileBuf bytes.Buffer
				if err := fileTemp.Execute(&fileBuf, payload); err != nil {
					log.Fatal(fmt.Errorf("executing file template: %w", err))
				}

				result.Files = append(result.Files, txtar.File{Name: fileBuf.String()})
			}
		}
	}

	result.Comment = caseBuf.Bytes()
	fmt.Println(string(txtar.Format(&result)))
}

func sliceOfAny[T any](t []T) []any {
	res := make([]any, len(t))
	for i, v := range t {
		res[i] = v
	}
	return res
}
