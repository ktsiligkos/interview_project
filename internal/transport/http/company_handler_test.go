package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ktsiligkos/xm_project/internal/domain"
	companyservice "github.com/ktsiligkos/xm_project/internal/service/company"
)

type stubCompanyService struct {
	getFn    func(ctx context.Context, companyID string) (domain.Company, error)
	createFn func(ctx context.Context, company domain.Company) (domain.Company, error)
	deleteFn func(ctx context.Context, companyID string) error
	patchFn  func(ctx context.Context, req domain.PatchCompanyRequest, uuid string) error
}

func (s stubCompanyService) GetCompanyByID(ctx context.Context, companyID string) (domain.Company, error) {
	if s.getFn == nil {
		return domain.Company{}, errors.New("unexpected call to GetCompanyByID")
	}
	return s.getFn(ctx, companyID)
}

func (s stubCompanyService) CreateCompany(ctx context.Context, company domain.Company) (domain.Company, error) {
	if s.createFn == nil {
		return domain.Company{}, errors.New("unexpected call to CreateCompany")
	}
	return s.createFn(ctx, company)
}

func (s stubCompanyService) DeleteCompanyByID(ctx context.Context, companyID string) error {
	if s.deleteFn == nil {
		return errors.New("unexpected call to DeleteCompanyByID")
	}
	return s.deleteFn(ctx, companyID)
}

func (s stubCompanyService) PatchCompanyByID(ctx context.Context, req domain.PatchCompanyRequest, uuid string) error {
	if s.patchFn == nil {
		return errors.New("unexpected call to PatchCompanyByID")
	}
	return s.patchFn(ctx, req, uuid)
}

func ptr[T any](v T) *T {
	return &v
}

func Given(t *testing.T, msg string, kv ...any) {
	t.Helper()
	if len(kv) > 0 {
		t.Logf("GIVEN: "+msg, kv...)
		return
	}
	t.Logf("GIVEN: %s", msg)
}

func When(t *testing.T, msg string, kv ...any) {
	t.Helper()
	if len(kv) > 0 {
		t.Logf("WHEN: "+msg, kv...)
		return
	}
	t.Logf("WHEN: %s", msg)
}

func Then(t *testing.T, msg string, kv ...any) {
	t.Helper()
	if len(kv) > 0 {
		t.Logf("THEN: "+msg, kv...)
		return
	}
	t.Logf("THEN: %s", msg)
}

// this function call the handlers with the correct method while injecting values in the Gin context
// (the body byte[] is used when there is a request with a body while, setup injects values when a url path paramerter is required)
// it returns the answer from the http recorder
func performRequest(t *testing.T, handler func(*gin.Context), method, target string, body []byte, setup func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequest(method, target, reader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.Request = req

	if setup != nil {
		setup(c)
	}

	handler(c)
	return w
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if got := w.Code; got != want {
		t.Fatalf("status code mismatch: got %d want %d; body=%s", got, want, strings.TrimSpace(w.Body.String()))
	}
}

func decodeBody[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("failed to unmarshal body %q: %v", w.Body.String(), err)
	}
	return out
}

func assertCompanyEqual(t *testing.T, got, want domain.Company) {
	t.Helper()
	if got.ID != want.ID {
		t.Fatalf("company id mismatch: got %q want %q", got.ID, want.ID)
	}
	if got.Name != want.Name {
		t.Fatalf("company name mismatch: got %q want %q", got.Name, want.Name)
	}
	if (got.Description == nil) != (want.Description == nil) {
		t.Fatalf("description nil mismatch: got=%v want=%v", got.Description, want.Description)
	}
	if got.Description != nil && want.Description != nil && *got.Description != *want.Description {
		t.Fatalf("description mismatch: got %q want %q", *got.Description, *want.Description)
	}
	if got.AmountOfEmployees != want.AmountOfEmployees {
		t.Fatalf("employees mismatch: got %d want %d", got.AmountOfEmployees, want.AmountOfEmployees)
	}
	if got.Registered != want.Registered {
		t.Fatalf("registered mismatch: got %v want %v", got.Registered, want.Registered)
	}
	if got.Type != want.Type {
		t.Fatalf("type mismatch: got %q want %q", got.Type, want.Type)
	}
}

func TestCompaniesHandler_Get_Success(t *testing.T) {
	Given(t, "an existing company id")

	expected := domain.Company{
		ID: "company-123", Name: "TechCorp",
		Description: ptr("A software company"), AmountOfEmployees: 100,
		Registered: true, Type: domain.Corporations,
	}
	service := stubCompanyService{
		getFn: func(_ context.Context, id string) (domain.Company, error) {
			if id != expected.ID {
				t.Fatalf("expected id %q, got %q", expected.ID, id)
			}
			return expected, nil
		},
	}
	handler := NewCompaniesHandler(service, nil)

	When(t, "GET /companies/:uuid is called")
	w := performRequest(t, handler.Get, http.MethodGet, "/companies/"+expected.ID, nil, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: expected.ID}}
	})

	Then(t, "it returns the company")
	assertStatus(t, w, http.StatusOK)
	got := decodeBody[domain.Company](t, w)
	assertCompanyEqual(t, got, expected)
}

func TestCompaniesHandler_Get_NotFound(t *testing.T) {
	Given(t, "a missing company id")

	service := stubCompanyService{
		getFn: func(context.Context, string) (domain.Company, error) {
			return domain.Company{}, companyservice.ErrNotFound
		},
	}
	handler := NewCompaniesHandler(service, nil)

	When(t, "GET /companies/:uuid is called")
	w := performRequest(t, handler.Get, http.MethodGet, "/companies/missing", nil, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "missing"}}
	})

	Then(t, "it returns not found")
	assertStatus(t, w, http.StatusNotFound)
	body := decodeBody[map[string]string](t, w)
	if got := body["error"]; got != "company not found" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Get_GenericError(t *testing.T) {
	Given(t, "a repo error while fetching a company")

	service := stubCompanyService{
		getFn: func(context.Context, string) (domain.Company, error) {
			return domain.Company{}, errors.New("db down")
		},
	}
	handler := NewCompaniesHandler(service, nil)

	When(t, "GET /companies/:uuid is called")
	w := performRequest(t, handler.Get, http.MethodGet, "/companies/err", nil, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "err"}}
	})

	Then(t, "it returns internal server error")
	assertStatus(t, w, http.StatusInternalServerError)
	body := decodeBody[map[string]string](t, w)
	if got := body["error"]; got != "failed to fetch company" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Create_InvalidBody(t *testing.T) {
	Given(t, "an invalid create request body")

	service := stubCompanyService{
		createFn: func(context.Context, domain.Company) (domain.Company, error) {
			t.Fatal("createFn should not be called on invalid payload")
			return domain.Company{}, nil
		},
	}
	handler := NewCompaniesHandler(service, nil)

	When(t, "POST /companies is called with malformed json")
	w := performRequest(t, handler.Create, http.MethodPost, "/companies", []byte("{invalid json"), nil)

	Then(t, "it returns bad request")
	assertStatus(t, w, http.StatusBadRequest)
	body := decodeBody[map[string]string](t, w)
	if got := body["error"]; got != "invalid request body" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Create_ValidationError(t *testing.T) {
	Given(t, "a create request that violates validation rules")

	service := stubCompanyService{
		createFn: func(context.Context, domain.Company) (domain.Company, error) {
			return domain.Company{}, fmt.Errorf("%w: name exceeds limit", companyservice.ErrValidationError)
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{
		"name":                "TooLongName",
		"amount_of_employees": 10,
		"registered":          true,
		"type":                domain.Corporations,
	}
	body, _ := json.Marshal(payload)

	When(t, "POST /companies is called")
	w := performRequest(t, handler.Create, http.MethodPost, "/companies", body, nil)

	Then(t, "it returns a validation error message")
	assertStatus(t, w, http.StatusBadRequest)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["error"]; got != "name exceeds limit" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Create_UniquenessViolation(t *testing.T) {
	Given(t, "a duplicate company name")

	service := stubCompanyService{
		createFn: func(context.Context, domain.Company) (domain.Company, error) {
			return domain.Company{}, fmt.Errorf("%w: name already exists", companyservice.ErrUniquenessViolation)
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{
		"name":                "Acme",
		"amount_of_employees": 10,
		"registered":          true,
		"type":                domain.Corporations,
	}
	body, _ := json.Marshal(payload)

	When(t, "POST /companies is called")
	w := performRequest(t, handler.Create, http.MethodPost, "/companies", body, nil)

	Then(t, "it returns bad request with the service error")
	assertStatus(t, w, http.StatusConflict)
	resp := decodeBody[map[string]string](t, w)
	want := fmt.Sprintf("%s: name already exists", companyservice.ErrUniquenessViolation)
	if got := resp["error"]; got != want {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Create_InvalidInput(t *testing.T) {
	Given(t, "a create call that triggers invalid input")

	service := stubCompanyService{
		createFn: func(context.Context, domain.Company) (domain.Company, error) {
			return domain.Company{}, companyservice.ErrInvalidInput
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{
		"name":                "Acme",
		"amount_of_employees": 10,
		"registered":          true,
		"type":                domain.Corporations,
	}
	body, _ := json.Marshal(payload)

	When(t, "POST /companies is called")
	w := performRequest(t, handler.Create, http.MethodPost, "/companies", body, nil)

	Then(t, "it returns bad request with the invalid input error")
	assertStatus(t, w, http.StatusBadRequest)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["error"]; got != companyservice.ErrInvalidInput.Error() {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Create_GenericError(t *testing.T) {
	Given(t, "a create call that fails unexpectedly")

	service := stubCompanyService{
		createFn: func(context.Context, domain.Company) (domain.Company, error) {
			return domain.Company{}, errors.New("db down")
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{
		"name":                "Acme",
		"amount_of_employees": 10,
		"registered":          true,
		"type":                domain.Corporations,
	}
	body, _ := json.Marshal(payload)

	When(t, "POST /companies is called")
	w := performRequest(t, handler.Create, http.MethodPost, "/companies", body, nil)

	Then(t, "it returns internal server error")
	assertStatus(t, w, http.StatusInternalServerError)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["error"]; got != "failed to create company" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Create_Success(t *testing.T) {
	Given(t, "a valid create request payload")

	expected := domain.Company{
		ID: "company-123", Name: "Acme",
		Description: ptr("desc"), AmountOfEmployees: 42,
		Registered: true, Type: domain.Corporations,
	}
	var captured domain.Company
	service := stubCompanyService{
		createFn: func(_ context.Context, company domain.Company) (domain.Company, error) {
			captured = company
			return expected, nil
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{
		"name":                "Acme",
		"description":         "desc",
		"amount_of_employees": 42,
		"registered":          true,
		"type":                domain.Corporations,
	}
	body, _ := json.Marshal(payload)

	When(t, "POST /companies is called")
	w := performRequest(t, handler.Create, http.MethodPost, "/companies", body, nil)

	Then(t, "it delegates to the service and returns the created company")
	if captured.Name != "Acme" || captured.AmountOfEmployees != 42 || !captured.Registered || captured.Type != domain.Corporations {
		t.Fatalf("unexpected payload sent to service: %+v", captured)
	}
	if captured.ID == "" {
		t.Fatalf("expected generated id to be non-empty")
	}

	assertStatus(t, w, http.StatusCreated)
	got := decodeBody[domain.Company](t, w)
	assertCompanyEqual(t, got, expected)
}

func TestCompaniesHandler_Patch_InvalidBody(t *testing.T) {
	Given(t, "an invalid patch request body")

	service := stubCompanyService{
		patchFn: func(context.Context, domain.PatchCompanyRequest, string) error {
			t.Fatal("patchFn should not be called on invalid payload")
			return nil
		},
	}
	handler := NewCompaniesHandler(service, nil)

	When(t, "PATCH /companies/:uuid is called with malformed json")
	w := performRequest(t, handler.Patch, http.MethodPatch, "/companies/company-123", []byte("{invalid"), func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "company-123"}}
	})

	Then(t, "it returns bad request")
	assertStatus(t, w, http.StatusBadRequest)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["error"]; got != "invalid request body" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Patch_ValidationError(t *testing.T) {
	Given(t, "a patch request that violates validation rules")

	service := stubCompanyService{
		patchFn: func(context.Context, domain.PatchCompanyRequest, string) error {
			return fmt.Errorf("%w: invalid name", companyservice.ErrValidationError)
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{"name": "TooLongName"}
	body, _ := json.Marshal(payload)

	When(t, "PATCH /companies/:uuid is called")
	w := performRequest(t, handler.Patch, http.MethodPatch, "/companies/company-123", body, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "company-123"}}
	})

	Then(t, "it returns validation error")
	assertStatus(t, w, http.StatusBadRequest)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["error"]; got != "invalid name" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Patch_NotFound(t *testing.T) {
	Given(t, "a patch attempt for a missing company")

	service := stubCompanyService{
		patchFn: func(context.Context, domain.PatchCompanyRequest, string) error {
			return companyservice.ErrNotFound
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{"name": "Acme"}
	body, _ := json.Marshal(payload)

	When(t, "PATCH /companies/:uuid is called")
	w := performRequest(t, handler.Patch, http.MethodPatch, "/companies/company-123", body, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "company-123"}}
	})

	Then(t, "it returns not found")
	assertStatus(t, w, http.StatusNotFound)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["error"]; got != "company not found" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Patch_UniquenessViolation(t *testing.T) {
	Given(t, "a patch that violates uniqueness constraints")

	service := stubCompanyService{
		patchFn: func(context.Context, domain.PatchCompanyRequest, string) error {
			return fmt.Errorf("%w: name already taken", companyservice.ErrUniquenessViolation)
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{"name": "Acme"}
	body, _ := json.Marshal(payload)

	When(t, "PATCH /companies/:uuid is called")
	w := performRequest(t, handler.Patch, http.MethodPatch, "/companies/company-123", body, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "company-123"}}
	})

	Then(t, "it returns bad request with the service error")
	assertStatus(t, w, http.StatusConflict)
	resp := decodeBody[map[string]string](t, w)
	want := fmt.Sprintf("%s: name already taken", companyservice.ErrUniquenessViolation)
	if got := resp["error"]; got != want {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Patch_InvalidInput(t *testing.T) {
	Given(t, "a patch that triggers invalid input")

	service := stubCompanyService{
		patchFn: func(context.Context, domain.PatchCompanyRequest, string) error {
			return companyservice.ErrInvalidInput
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{"name": "Acme"}
	body, _ := json.Marshal(payload)

	When(t, "PATCH /companies/:uuid is called")
	w := performRequest(t, handler.Patch, http.MethodPatch, "/companies/company-123", body, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "company-123"}}
	})

	Then(t, "it returns bad request with the invalid input error")
	assertStatus(t, w, http.StatusBadRequest)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["error"]; got != companyservice.ErrInvalidInput.Error() {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Patch_GenericError(t *testing.T) {
	Given(t, "a patch that fails unexpectedly")

	service := stubCompanyService{
		patchFn: func(context.Context, domain.PatchCompanyRequest, string) error {
			return errors.New("db blew up")
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{"name": "Acme"}
	body, _ := json.Marshal(payload)

	When(t, "PATCH /companies/:uuid is called")
	w := performRequest(t, handler.Patch, http.MethodPatch, "/companies/company-123", body, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "company-123"}}
	})

	Then(t, "it returns internal server error")
	assertStatus(t, w, http.StatusInternalServerError)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["error"]; got != "failed to update company" {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func TestCompaniesHandler_Patch_Success(t *testing.T) {
	Given(t, "a valid patch request")

	var captured domain.PatchCompanyRequest
	var capturedID string
	service := stubCompanyService{
		patchFn: func(_ context.Context, req domain.PatchCompanyRequest, id string) error {
			captured = req
			capturedID = id
			return nil
		},
	}
	handler := NewCompaniesHandler(service, nil)

	payload := map[string]any{"name": "Acme"}
	body, _ := json.Marshal(payload)

	When(t, "PATCH /companies/:uuid is called")
	w := performRequest(t, handler.Patch, http.MethodPatch, "/companies/company-123", body, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "company-123"}}
	})

	Then(t, "it calls the service and returns success")
	if captured.Name == nil || *captured.Name != "Acme" {
		t.Fatalf("unexpected patch payload: %+v", captured)
	}
	if capturedID != "company-123" {
		t.Fatalf("unexpected patch id: %q", capturedID)
	}

	assertStatus(t, w, http.StatusCreated)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["status"]; got != "success" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestCompaniesHandler_Delete_Success(t *testing.T) {
	Given(t, "an existing company id for deletion")

	var capturedID string
	service := stubCompanyService{
		deleteFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	handler := NewCompaniesHandler(service, nil)

	When(t, "DELETE /companies/:uuid is called")
	w := performRequest(t, handler.Delete, http.MethodDelete, "/companies/company-123", nil, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "company-123"}}
	})

	Then(t, "it calls the service and returns success")
	if capturedID != "company-123" {
		t.Fatalf("unexpected delete id: %q", capturedID)
	}
	assertStatus(t, w, http.StatusOK)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["status"]; got != "success" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestCompaniesHandler_Delete_NotFound(t *testing.T) {
	Given(t, "a delete call for a missing company")

	service := stubCompanyService{
		deleteFn: func(context.Context, string) error {
			return companyservice.ErrNotFound
		},
	}
	handler := NewCompaniesHandler(service, nil)

	When(t, "DELETE /companies/:uuid is called")
	w := performRequest(t, handler.Delete, http.MethodDelete, "/companies/company-123", nil, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "company-123"}}
	})

	Then(t, "it returns not found")
	assertStatus(t, w, http.StatusNotFound)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["error"]; got != "company not found" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestCompaniesHandler_Delete_GenericError(t *testing.T) {
	Given(t, "a delete call that fails unexpectedly")

	service := stubCompanyService{
		deleteFn: func(context.Context, string) error {
			return errors.New("db down")
		},
	}
	handler := NewCompaniesHandler(service, nil)

	When(t, "DELETE /companies/:uuid is called")
	w := performRequest(t, handler.Delete, http.MethodDelete, "/companies/company-123", nil, func(c *gin.Context) {
		c.Params = gin.Params{gin.Param{Key: "uuid", Value: "company-123"}}
	})

	Then(t, "it returns internal server error")
	assertStatus(t, w, http.StatusInternalServerError)
	resp := decodeBody[map[string]string](t, w)
	if got := resp["error"]; got != "failed to delete the company" {
		t.Fatalf("unexpected response: %v", resp)
	}
}
