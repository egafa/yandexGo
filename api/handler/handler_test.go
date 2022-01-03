package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/egafa/yandexGo/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetricHandlerChi(t *testing.T) {
	type want struct {
		response   string
		statusCode int
	}

	model.InitMapMetricVal()

	tests := []struct {
		name    string
		request string
		want    want
	}{
		{name: "Update Metric test",
			request: "http://127.0.0.1:8080/update/counter/Sys/22",
			want: want{
				response:   "22",
				statusCode: 200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateMetricHandlerChi)
			h.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			response, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.response, response)

		})
	}
}
