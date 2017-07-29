package apirest

import (
	"fmt"
	"github.com/fatih/color"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/RangelReale/osin"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/RichardKnop/machinery/v1"
	"github.com/garyburd/redigo/redis"
	"github.com/ShaleApps/osinredis"
	"github.com/golang/protobuf/proto"

	"github.com/ulule/deepcopier"
	"io/ioutil"
	"gopkg.in/go-playground/validator.v9"
	"github.com/maxpowel/dislet"
	"github.com/maxpowel/dislet/apirest/protomodel"
	mprotomodel "github.com/maxpowel/dislet/machinery/protomodel"

)

type Config struct {
	Port int
}


// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (se StatusError) Error() string {
	return se.Err.Error()
}

// Returns our HTTP status code.
func (se StatusError) Status() int {
	return se.Code
}

type Handler struct {
	*dislet.Kernel
	H func(k *dislet.Kernel, w http.ResponseWriter, r *http.Request) error
}

func GetBody(protoMessage proto.Message, r *http.Request) (error){
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return proto.Unmarshal(buf, protoMessage)

}
// ServeHTTP allows our Handler type to satisfy http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(h.Kernel, w, r)
	if err != nil {
		switch e := err.(type) {
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			errorProto := &protomodel.Error{
				Code: int32(e.Status()),
				Description: e.Error(),
			}

			data, err := proto.Marshal(errorProto)

			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
			}

			w.WriteHeader(http.StatusBadRequest)
			// Raw binary data is sent
			w.Write(data)
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}


// Format task information. Used everytime your controller runs a task
func TaskResponseHandler(result *tasks.TaskState) ([]byte, error){
	state := mprotomodel.TaskState_UNKWNOWN

	switch result.State {
	case "PENDING": state = mprotomodel.TaskState_PENDING
	case "RECEIVED": state = mprotomodel.TaskState_RECEIVED
	case "STARTED": state = mprotomodel.TaskState_STARTED
	case "RETRY": state = mprotomodel.TaskState_RETRY
	case "SUCCESS": state = mprotomodel.TaskState_SUCCESS
	case "FAILURE": state = mprotomodel.TaskState_FAILURE
	}


	ts := mprotomodel.TaskStateResponse{
		State: state,
		ETA: 0,
		Uid: result.TaskUUID,
	}

	if len(result.Results) > 1 {
		if result.Results[0].Type == "string"  && result.Results[0].Value != nil && result.Results[1].Type == "map[string]string"  && result.Results[1].Value != nil{
			m := make(map[string]string)
			for k, v := range result.Results[1].Value.(map[string]interface{}) {
				m[k] = v.(string)
			}
			taskError := mprotomodel.TaskError{
				Code: 250,
				Format: result.Results[0].Value.(string),
				Params: m,
			}
			ts.Error = &taskError

		}
	}

		fmt.Println(result)
		/**/


	return proto.Marshal(&ts)
}

// Shortcut to launch a task
func SendTask(kernel *dislet.Kernel, task *tasks.Signature) ([]byte, error){
	server := kernel.Container.MustGet("machinery").(*machinery.Server)
	asyncResult, err := server.SendTask(task)
	if err != nil {
		return nil, err
	}

	return TaskResponseHandler(asyncResult.GetState())
}


// Validate input data against a model
func Validate(data interface{}, validatorI interface{}) (*interface{}, error) {
	var validate *validator.Validate
	validate = validator.New()


	deepcopier.Copy(data).To(validatorI)
	err := validate.Struct(validatorI)
	//_, err := govalidator.ValidateStruct(validator)
	return &validatorI, err
}

// TODO MOVER a un sitio correcto
func NewRedisStorage() (*osinredis.Storage){
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", ":6379")
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}

	storage := osinredis.New(pool, "prefix")
	storage.CreateClient(&osin.DefaultClient{
		Id: "pepe",
		RedirectUri: "http://google.es",
		Secret: "lolazo",
	})
	return storage
}

func NewOAuthServer(k *dislet.Kernel) *osin.Server {
	oauthConfig := osin.NewServerConfig()
	oauthConfig.AllowedAccessTypes = osin.AllowedAccessType{osin.PASSWORD}
	return osin.NewServer(oauthConfig, NewRedisStorage())
}

func NewApiRest(k *dislet.Kernel, port int) *mux.Router {

	router := mux.NewRouter().StrictSlash(true)


	k.Container.RegisterType("oauth", NewOAuthServer, k)
	k.Container.MustGet("oauth")


	// Authorization code endpoint
	/*router.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		defer resp.Close()

		if ar := server.HandleAuthorizeRequest(resp, r); ar != nil {

			// HANDLE LOGIN PAGE HERE

			ar.Authorized = true
			server.FinishAuthorizeRequest(resp, r, ar)
		}
		osin.OutputJSON(resp, w, r)
	})*/

	/*router.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		defer resp.Close()

		if ir := server.HandleInfoRequest(resp, r); ir != nil {
			fmt.Println("AA")
			server.FinishInfoRequest(resp, r, ir)
			fmt.Println("B")
		}
		o := osin.ResponseData{}
		o["lol"] = "lel"
		resp.Output = o
		osin.OutputJSON(resp, w, r)
	})*/

//authorize?response_type=code&client_id=1234&redirect_uri=http%3A%2F%2Flocalhost%3A14000%2Fappauth%2Fcode
//curl 'http://localhost:8090/token' -d 'grant_type=password&username=pepe&password=21212&client_id=pepe' -H 'Authorization: Basic cGVwZTpsb2xhem8='
	// Access token endpoint
	//router.Handle("/token", Handler{k, CheckToken})

	go http.ListenAndServe(fmt.Sprintf(":%v", port), router)
	color.Cyan("Api listening on port %v", port)

	return router
}


func Bootstrap(k *dislet.Kernel) {
	//fmt.Println("DATABASE BOOT")
	mapping := k.Config.Mapping
	mapping["api"] = &Config{}

	var baz dislet.OnKernelReady = func(k *dislet.Kernel){
		color.Green("Booting api")
		conf := k.Config.Mapping["api"].(*Config)
		k.Container.RegisterType("api", NewApiRest, k, conf.Port)
		k.Container.MustGet("api")


	}
	k.Subscribe(baz)




}
