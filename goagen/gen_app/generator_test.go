package genapp_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_app"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate", func() {
	var workspace *codegen.Workspace
	var outDir string
	var files []string
	var genErr error

	BeforeEach(func() {
		var err error
		workspace, err = codegen.NewWorkspace("test")
		Ω(err).ShouldNot(HaveOccurred())
		outDir, err = ioutil.TempDir(filepath.Join(workspace.Path, "src"), "")
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"goagen", "--out=" + outDir, "--design=foo"}
	})

	JustBeforeEach(func() {
		files, genErr = genapp.Generate([]interface{}{design.Design})
	})

	AfterEach(func() {
		workspace.Delete()
	})

	Context("with a dummy API", func() {
		BeforeEach(func() {
			design.Design = &design.APIDefinition{
				APIVersionDefinition: &design.APIVersionDefinition{
					Name:        "test api",
					Title:       "dummy API with no resource",
					Description: "I told you it's dummy",
				},
			}
		})

		It("generates correct empty files", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(6))
			isEmptySource := func(filename string) {
				contextsContent, err := ioutil.ReadFile(filepath.Join(outDir, "app", filename))
				Ω(err).ShouldNot(HaveOccurred())
				lines := strings.Split(string(contextsContent), "\n")
				Ω(lines).ShouldNot(BeEmpty())
				Ω(len(lines)).Should(BeNumerically(">", 1))
			}
			isEmptySource("contexts.go")
			isEmptySource("controllers.go")
			isEmptySource("hrefs.go")
			isEmptySource("media_types.go")
		})
	})

	Context("with a simple API", func() {
		var contextsCode, controllersCode, hrefsCode, mediaTypesCode, version string
		var payload *design.UserTypeDefinition

		isSource := func(filename, content string) {
			contextsContent, err := ioutil.ReadFile(filepath.Join(outDir, "app", filename))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(string(contextsContent)).Should(Equal(content))
		}

		funcs := template.FuncMap{
			"sep": func() string { return string(os.PathSeparator) },
		}

		runCodeTemplates := func(data map[string]string) {
			contextsCodeT, err := template.New("context").Funcs(funcs).Parse(contextsCodeTmpl)
			Ω(err).ShouldNot(HaveOccurred())
			var b bytes.Buffer
			err = contextsCodeT.Execute(&b, data)
			Ω(err).ShouldNot(HaveOccurred())
			contextsCode = b.String()

			controllersCodeT, err := template.New("controllers").Funcs(funcs).Parse(controllersCodeTmpl)
			Ω(err).ShouldNot(HaveOccurred())
			b.Reset()
			err = controllersCodeT.Execute(&b, data)
			Ω(err).ShouldNot(HaveOccurred())
			controllersCode = b.String()

			hrefsCodeT, err := template.New("hrefs").Funcs(funcs).Parse(hrefsCodeTmpl)
			Ω(err).ShouldNot(HaveOccurred())
			b.Reset()
			err = hrefsCodeT.Execute(&b, data)
			Ω(err).ShouldNot(HaveOccurred())
			hrefsCode = b.String()

			mediaTypesCodeT, err := template.New("media types").Funcs(funcs).Parse(mediaTypesCodeTmpl)
			Ω(err).ShouldNot(HaveOccurred())
			b.Reset()
			err = mediaTypesCodeT.Execute(&b, data)
			Ω(err).ShouldNot(HaveOccurred())
			mediaTypesCode = b.String()
		}

		BeforeEach(func() {
			payload = nil
			required := &dslengine.ValidationDefinition{
				Required: []string{"id"},
			}
			idAt := design.AttributeDefinition{
				Type:        design.String,
				Description: "widget id",
			}
			params := design.AttributeDefinition{
				Type: design.Object{
					"id": &idAt,
				},
				Validation: required,
			}
			resp := design.ResponseDefinition{
				Name:        "ok",
				Status:      200,
				Description: "get of widgets",
				MediaType:   "vnd.rightscale.codegen.test.widgets",
			}
			route := design.RouteDefinition{
				Verb: "GET",
				Path: "/:id",
			}
			at := design.AttributeDefinition{
				Type: design.String,
			}
			ut := design.UserTypeDefinition{
				AttributeDefinition: &at,
				TypeName:            "id",
			}
			res := design.ResourceDefinition{
				Name:                "Widget",
				BasePath:            "/widgets",
				Description:         "Widgetty",
				MediaType:           "vnd.rightscale.codegen.test.widgets",
				CanonicalActionName: "get",
			}
			get := design.ActionDefinition{
				Name:        "get",
				Description: "get widgets",
				Parent:      &res,
				Routes:      []*design.RouteDefinition{&route},
				Responses:   map[string]*design.ResponseDefinition{"ok": &resp},
				Params:      &params,
				Payload:     payload,
			}
			res.Actions = map[string]*design.ActionDefinition{"get": &get}
			mt := design.MediaTypeDefinition{
				UserTypeDefinition: &ut,
				Identifier:         "vnd.rightscale.codegen.test.widgets",
				Views: map[string]*design.ViewDefinition{
					"default": {
						AttributeDefinition: ut.AttributeDefinition,
						Name:                "default",
					},
				},
			}
			design.Design = &design.APIDefinition{
				APIVersionDefinition: &design.APIVersionDefinition{
					Name:        "test api",
					Title:       "dummy API with no resource",
					Description: "I told you it's dummy",
				},
				Resources:  map[string]*design.ResourceDefinition{"Widget": &res},
				MediaTypes: map[string]*design.MediaTypeDefinition{"vnd.rightscale.codegen.test.widgets": &mt},
			}
		})

		Context("", func() {
			BeforeEach(func() {
				runCodeTemplates(map[string]string{"outDir": outDir, "design": "foo", "version": "", "tmpDir": filepath.Base(outDir)})
			})

			It("generates the corresponding code", func() {
				Ω(genErr).Should(BeNil())
				Ω(files).Should(HaveLen(6))

				isSource("contexts.go", contextsCode)
				isSource("controllers.go", controllersCode)
				isSource("hrefs.go", hrefsCode)
				isSource("media_types.go", mediaTypesCode)
			})
		})

		Context("that is versioned", func() {
			BeforeEach(func() {
				version = "v1"
				design.Design.APIVersions = make(map[string]*design.APIVersionDefinition)
				verDef := &design.APIVersionDefinition{}
				verDef.Version = version
				design.Design.APIVersions[version] = verDef
				design.Design.Resources["Widget"].APIVersions = []string{version}
				runCodeTemplates(map[string]string{
					"outDir":  outDir,
					"tmpDir":  filepath.Base(outDir),
					"design":  "foo",
					"version": version,
				})
			})

			It("generates the versioned code", func() {
				Ω(genErr).Should(BeNil())
				Ω(files).Should(HaveLen(11))

				isSource(version+"/contexts.go", contextsCode)
				isSource(version+"/controllers.go", controllersCode)
				isSource(version+"/hrefs.go", hrefsCode)
				isSource("media_types.go", mediaTypesCode)
			})
		})

		Context("with a slice payload", func() {
			BeforeEach(func() {
				elemType := &design.AttributeDefinition{Type: design.Integer}
				payload = &design.UserTypeDefinition{
					AttributeDefinition: &design.AttributeDefinition{
						Type: &design.Array{ElemType: elemType},
					},
					TypeName: "Collection",
				}
				design.Design.Resources["Widget"].Actions["get"].Payload = payload
				runCodeTemplates(map[string]string{"outDir": outDir, "design": "foo", "version": "", "tmpDir": filepath.Base(outDir)})
			})

			It("generates the correct payload assignment code", func() {
				Ω(genErr).Should(BeNil())

				contextsContent, err := ioutil.ReadFile(filepath.Join(outDir, "app", "controllers.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(contextsContent)).Should(ContainSubstring(controllersSlicePayloadCode))
			})
		})

	})
})

var _ = Describe("BuildEncoderMap", func() {
	var info []*design.EncodingDefinition
	var encoder bool

	var data map[string]*genapp.EncoderTemplateData
	var resErr error

	BeforeEach(func() {
		info = nil
		encoder = false
	})

	JustBeforeEach(func() {
		data, resErr = genapp.BuildEncoderMap(info, encoder)
	})

	Context("with a single definition using a single known MIME type for encoding", func() {
		BeforeEach(func() {
			simple := &design.EncodingDefinition{
				MIMETypes: []string{"application/json"},
			}
			info = append(info, simple)
			encoder = true
		})

		It("generates a map with a single entry", func() {
			Ω(resErr).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveLen(1))
			Ω(data).Should(HaveKey("json"))
			jd := data["json"]
			Ω(jd).ShouldNot(BeNil())
			Ω(jd.PackagePath).Should(Equal("json"))
			Ω(jd.PackageName).Should(Equal("goa"))
			Ω(jd.Factory).Should(Equal("JSONEncoderFactory"))
			Ω(jd.MIMETypes).Should(HaveLen(1))
			Ω(jd.MIMETypes[0]).Should(Equal("application/json"))
		})
	})

	Context("with a single definition using a single known MIME type for decoding", func() {
		BeforeEach(func() {
			simple := &design.EncodingDefinition{
				MIMETypes: []string{"application/json"},
			}
			info = append(info, simple)
			encoder = false
		})

		It("generates a map with a single entry", func() {
			Ω(resErr).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveLen(1))
			Ω(data).Should(HaveKey("json"))
			jd := data["json"]
			Ω(jd).ShouldNot(BeNil())
			Ω(jd.PackagePath).Should(Equal("json"))
			Ω(jd.PackageName).Should(Equal("goa"))
			Ω(jd.Factory).Should(Equal("JSONDecoderFactory"))
			Ω(jd.MIMETypes).Should(HaveLen(1))
			Ω(jd.MIMETypes[0]).Should(Equal("application/json"))
		})
	})

	Context("with a definition using a custom decoding package", func() {
		const packagePath = "github.com/goadesign/goa/design" // Just to pick something always available
		var mimeTypes = []string{"application/vnd.custom", "application/vnd.custom2"}

		BeforeEach(func() {
			simple := &design.EncodingDefinition{
				PackagePath: packagePath,
				MIMETypes:   mimeTypes,
			}
			info = append(info, simple)
		})

		It("generates a map with a single entry", func() {
			Ω(resErr).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveLen(1))
			Ω(data).Should(HaveKey(packagePath))
			jd := data[packagePath]
			Ω(jd).ShouldNot(BeNil())
			Ω(jd.PackagePath).Should(Equal(packagePath))
			Ω(jd.PackageName).Should(Equal("design"))
			Ω(jd.Factory).Should(Equal("DecoderFactory"))
			Ω(jd.MIMETypes).Should(ConsistOf(interface{}(mimeTypes[0]), interface{}(mimeTypes[1])))
		})
	})

	Context("with a definition using a custom decoding package for a known encoding", func() {
		const packagePath = "github.com/goadesign/goa/design" // Just to pick something always available
		var mimeTypes = []string{"application/json"}

		BeforeEach(func() {
			simple := &design.EncodingDefinition{
				PackagePath: packagePath,
				MIMETypes:   mimeTypes,
			}
			info = append(info, simple)
		})

		It("generates a map with a single entry using the generic decoder factory name", func() {
			Ω(resErr).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveLen(1))
			Ω(data).Should(HaveKey(packagePath))
			jd := data[packagePath]
			Ω(jd).ShouldNot(BeNil())
			Ω(jd.PackagePath).Should(Equal(packagePath))
			Ω(jd.PackageName).Should(Equal("design"))
			Ω(jd.Factory).Should(Equal("DecoderFactory"))
		})

	})
})

const contextsCodeTmpl = `//************************************************************************//
// API "test api"{{if .version}} version {{.version}}{{end}}: Application Contexts
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH){{sep}}src{{sep}}{{.tmpDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package {{if .version}}{{.version}}{{else}}app{{end}}

import (
{{if .version}}	"{{.tmpDir}}/app"
{{end}}	"github.com/goadesign/goa"
	"golang.org/x/net/context"
)

// GetWidgetContext provides the Widget get action context.
type GetWidgetContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData{{if .version}}
	ID         string
	APIVersion string{{else}}
	ID string{{end}}
}

// NewGetWidgetContext parses the incoming request URL and body, performs validations and creates the
// context used by the Widget controller get action.
func NewGetWidgetContext(ctx context.Context) (*GetWidgetContext, error) {
	var err error
	req := goa.Request(ctx)
	rctx := GetWidgetContext{Context: ctx, ResponseData: goa.Response(ctx), RequestData: req}
	rawID := req.Params.Get("id")
	if rawID != "" {
		rctx.ID = rawID
	}
	return &rctx, err
}

// OK sends a HTTP response with status code 200.
func (ctx *GetWidgetContext) OK(r {{if .version}}app.{{end}}ID) error {
	ctx.ResponseData.Header().Set("Content-Type", "vnd.rightscale.codegen.test.widgets")
	return ctx.ResponseData.Send(ctx.Context, 200, r)
}
`

const controllersCodeTmpl = `//************************************************************************//
// API "test api"{{if .version}} version {{.version}}{{end}}: Application Controllers
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH){{sep}}src{{sep}}{{.tmpDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package {{if .version}}{{.version}}{{else}}app{{end}}

import (
	"github.com/goadesign/goa"
	"golang.org/x/net/context"
	"net/http"
)

// inited is true if initService has been called
var inited = false

// initService sets up the service encoders, decoders and mux.
func initService(service *goa.Service) {
	if inited {
		return
	}
	inited = true
	// Setup encoders and decoders

}

// WidgetController is the controller interface for the Widget actions.
type WidgetController interface {
	goa.Muxer
	Get(*GetWidgetContext) error
}

// MountWidgetController "mounts" a Widget resource controller on the given service.
func MountWidgetController(service *goa.Service, ctrl WidgetController) {
	initService(service)
	var h goa.Handler
	mux := service.{{if .version}}Version("{{.version}}").Mux{{else}}Mux{{end}}
	h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		rctx, err := NewGetWidgetContext(ctx)
		if err != nil {
			return goa.NewBadRequestError(err)
		}{{if .version}}
		rctx.APIVersion = "{{.version}}"{{end}}
		return ctrl.Get(rctx)
	}
	mux.Handle("GET", "/:id", ctrl.MuxHandler("Get", h, nil))
	goa.Info(goa.RootContext, "mount", goa.KV{"ctrl", "Widget"},{{if .version}} goa.KV{"version", "{{.version}}"},{{end}} goa.KV{"action", "Get"}, goa.KV{"route", "GET /:id"})
}
`

const hrefsCodeTmpl = `//************************************************************************//
// API "test api"{{if .version}} version {{.version}}{{end}}: Application Resource Href Factories
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH){{sep}}src{{sep}}{{.tmpDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package {{if .version}}{{.version}}{{else}}app{{end}}

import "fmt"

// WidgetHref returns the resource href.
func WidgetHref(id interface{}) string {
	return fmt.Sprintf("/%v", id)
}
`

const mediaTypesCodeTmpl = `//************************************************************************//
// API "test api": Application Media Types
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH){{sep}}src{{sep}}{{.tmpDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app
`

const controllersSlicePayloadCode = `
// MountWidgetController "mounts" a Widget resource controller on the given service.
func MountWidgetController(service *goa.Service, ctrl WidgetController) {
	initService(service)
	var h goa.Handler
	mux := service.Mux
	h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		rctx, err := NewGetWidgetContext(ctx)
		if err != nil {
			return goa.NewBadRequestError(err)
		}
		if rawPayload := goa.Request(ctx).Payload; rawPayload != nil {
			rctx.Payload = rawPayload.(Collection)
		}
		return ctrl.Get(rctx)
	}
	mux.Handle("GET", "/:id", ctrl.MuxHandler("Get", h, unmarshalGetWidgetPayload))
	goa.Info(goa.RootContext, "mount", goa.KV{"ctrl", "Widget"}, goa.KV{"action", "Get"}, goa.KV{"route", "GET /:id"})
}

// unmarshalGetWidgetPayload unmarshals the request body into the context request data Payload field.
func unmarshalGetWidgetPayload(ctx context.Context, req *http.Request) error {
	var payload Collection
	if err := goa.RequestService(ctx).DecodeRequest(req, &payload); err != nil {
		return err
	}
	goa.Request(ctx).Payload = payload
	return nil
}
`
