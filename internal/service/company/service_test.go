package company

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ktsiligkos/xm_project/internal/domain"
	repoerrors "github.com/ktsiligkos/xm_project/internal/repository/company"
)

type stubRepository struct {
	getFn    func(ctx context.Context, companyID string) (domain.Company, error)
	createFn func(ctx context.Context, company domain.Company) (domain.Company, error)
	deleteFn func(ctx context.Context, companyID string) error
	patchFn  func(ctx context.Context, req domain.PatchCompanyRequest, uuid string, maxNumOfFields int) error
}

func (s stubRepository) GetCompanyByID(ctx context.Context, companyID string) (domain.Company, error) {
	if s.getFn != nil {
		return s.getFn(ctx, companyID)
	}
	return domain.Company{}, errors.New("unexpected call to GetCompanyByID")
}

func (s stubRepository) CreateCompany(ctx context.Context, company domain.Company) (domain.Company, error) {
	if s.createFn != nil {
		return s.createFn(ctx, company)
	}
	return domain.Company{}, errors.New("unexpected call to CreateCompany")
}

func (s stubRepository) DeleteCompanyByID(ctx context.Context, companyID string) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, companyID)
	}
	return errors.New("unexpected call to DeleteCompanyByID")
}

func (s stubRepository) PatchCompanyByID(ctx context.Context, req domain.PatchCompanyRequest, uuid string, maxNumOfFields int) error {
	if s.patchFn != nil {
		return s.patchFn(ctx, req, uuid, maxNumOfFields)
	}
	return errors.New("unexpected call to PatchCompanyByID")
}

// Holds the number of published events
type stubPublisher struct {
	events []CompanyEvent
	err    error
}

func (s *stubPublisher) PublishCompanyEvent(ctx context.Context, event CompanyEvent) error {
	s.events = append(s.events, event)
	return s.err
}

// helper function for passing a pointer to the struct field
func ptr[T any](v T) *T {
	return &v
}

// func TestCreateCompany_Success_Publish(t *testing.T) {
// 	expected := domain.Company{
// 		ID:                "company-123",
// 		Name:              "TechCorp",
// 		Description:       ptr("A software company"),
// 		AmountOfEmployees: 100,
// 		Registered:        true,
// 		Type:              domain.Corporations,
// 	}

// 	repo := stubRepository{
// 		createFn: func(ctx context.Context, company domain.Company) (domain.Company, error) {
// 			if company.Name != expected.Name {
// 				t.Fatalf("expected name %q but got %q", expected.Name, company.Name)
// 			}
// 			return expected, nil
// 		},
// 	}

// 	publisher := &stubPublisher{}
// 	svc := NewService(repo, publisher)

// 	got, err := svc.CreateCompany(context.Background(), expected)
// 	if err != nil {
// 		t.Fatalf("CreateCompany returned unexpected error: %v", err)
// 	}

// 	if got != expected {
// 		t.Fatalf("CreateCompany returned %+v, expected %+v", got, expected)
// 	}

// 	if len(publisher.events) != 1 {
// 		t.Fatalf("expected 1 event, got %d", len(publisher.events))
// 	}

// 	event := publisher.events[0]
// 	if event.Operation != "company.created" {
// 		t.Fatalf("expected operation company.created, got %s", event.Operation)
// 	}
// 	if event.Company.ID != expected.ID || event.Company.Name != expected.Name {
// 		t.Fatalf("unexpected event company payload: %+v", event.Company)
// 	}
// }

// func TestCreateCompany_ValidationError_NoPublish(t *testing.T) {
// 	repo := stubRepository{
// 		createFn: func(ctx context.Context, company domain.Company) (domain.Company, error) {
// 			t.Fatal("createFn should not be called when validation fails")
// 			return domain.Company{}, nil
// 		},
// 	}

// 	publisher := &stubPublisher{}
// 	svc := NewService(repo, publisher)
// 	longName := "ThisNameIsWayTooLong"
// 	company := domain.Company{
// 		ID:                "id",
// 		Name:              longName,
// 		Description:       ptr("desc"),
// 		AmountOfEmployees: 10,
// 		Registered:        true,
// 		Type:              domain.Corporations,
// 	}

// 	_, err := svc.CreateCompany(context.Background(), company)
// 	if err == nil {
// 		t.Fatal("expected validation error but got nil")
// 	}

// 	if !errors.Is(err, ErrValidationError) {
// 		t.Fatalf("expected ErrValidationError, got %v", err)
// 	}

// 	if len(publisher.events) != 0 {
// 		t.Fatalf("expected no events on validation failure, got %d", len(publisher.events))
// 	}
// }

// func TestCreateCompany_UniquenessViolation_NoPublish(t *testing.T) {
// 	repoErr := repoerrors.ErrUniquenessViolation
// 	repo := stubRepository{
// 		createFn: func(ctx context.Context, company domain.Company) (domain.Company, error) {
// 			return domain.Company{}, repoErr
// 		},
// 	}

// 	publisher := &stubPublisher{}
// 	svc := NewService(repo, publisher)
// 	company := domain.Company{
// 		ID:                "id",
// 		Name:              "UniqueName",
// 		Description:       ptr("desc"),
// 		AmountOfEmployees: 10,
// 		Registered:        true,
// 		Type:              domain.Corporations,
// 	}

// 	_, err := svc.CreateCompany(context.Background(), company)
// 	if err == nil {
// 		t.Fatal("expected uniqueness violation error but got nil")
// 	}

// 	if !errors.Is(err, ErrUniquenessViolation) {
// 		t.Fatalf("expected ErrUniquenessViolation, got %v", err)
// 	}

// 	if len(publisher.events) != 0 {
// 		t.Fatalf("expected no events on uniqueness violation, got %d", len(publisher.events))
// 	}
// }

// func TestCreateCompany_RepoGenericError_Bubbles_NoPublish(t *testing.T) {
// 	input := domain.Company{
// 		ID:   "company-1",
// 		Name: "TechCorp",
// 		Type: domain.Corporations,
// 	}

// 	genericError := errors.New("db down")
// 	repo := stubRepository{
// 		createFn: func(ctx context.Context, c domain.Company) (domain.Company, error) {
// 			return domain.Company{}, genericError
// 		},
// 	}
// 	pub := &stubPublisher{}
// 	svc := NewService(repo, pub)

// 	_, err := svc.CreateCompany(context.Background(), input)
// 	if err == nil || !errors.Is(err, genericError) {
// 		t.Fatalf("want original repo error, got %v", err)
// 	}
// 	if len(pub.events) != 0 {
// 		t.Fatalf("no event should be published when create fails")
// 	}
// }

// func TestGetCompanyByID_Success(t *testing.T) {
// 	want := domain.Company{
// 		ID:                "company-123",
// 		Name:              "TechCorp",
// 		Description:       ptr("A software company"),
// 		AmountOfEmployees: 100,
// 		Registered:        true,
// 		Type:              domain.Corporations,
// 	}

// 	repo := stubRepository{
// 		getFn: func(ctx context.Context, id string) (domain.Company, error) {
// 			if id != want.ID {
// 				t.Fatalf("repo received id=%q, want %q", id, want.ID)
// 			}
// 			return want, nil
// 		},
// 	}
// 	svc := NewService(repo, &stubPublisher{})

// 	got, err := svc.GetCompanyByID(context.Background(), want.ID)
// 	if err != nil {
// 		t.Fatalf("GetCompanyByID returned error: %v", err)
// 	}
// 	if !reflect.DeepEqual(got, want) {
// 		t.Fatalf("GetCompanyByID mismatch:\n got=%+v\nwant=%+v", got, want)
// 	}
// }

// func TestGetCompanyByID_NotFound(t *testing.T) {
// 	repo := stubRepository{
// 		getFn: func(ctx context.Context, companyID string) (domain.Company, error) {
// 			return domain.Company{}, repoerrors.ErrNotFound
// 		},
// 	}

// 	svc := NewService(repo, nil)

// 	_, err := svc.GetCompanyByID(context.Background(), "missing-id")
// 	if err == nil {
// 		t.Fatal("expected ErrNotFound but got nil")
// 	}

// 	if !errors.Is(err, ErrNotFound) {
// 		t.Fatalf("expected ErrNotFound, got %v", err)
// 	}
// }

// func TestGetCompanyByID_GenericRepoError_BubblesUp(t *testing.T) {
// 	genericError := errors.New("db connection failed")
// 	repo := stubRepository{
// 		getFn: func(ctx context.Context, id string) (domain.Company, error) {
// 			return domain.Company{}, genericError
// 		},
// 	}
// 	svc := NewService(repo, &stubPublisher{})

// 	_, err := svc.GetCompanyByID(context.Background(), "any-id")
// 	if err == nil {
// 		t.Fatalf("expected error, got nil")
// 	}
// 	// Should not map to ErrNotFound; it should be the original error
// 	if !errors.Is(err, genericError) {
// 		t.Fatalf("want original repo error %q, got %v", genericError, err)
// 	}
// }

// func TestDeleteCompanyByID_Success_Publish(t *testing.T) {
// 	const id = "company-123"

// 	repo := stubRepository{
// 		deleteFn: func(ctx context.Context, got string) error {
// 			if got != id {
// 				t.Fatalf("repo received id %q, want %q", got, id)
// 			}
// 			return nil
// 		},
// 	}
// 	pub := &stubPublisher{}
// 	svc := NewService(repo, pub)

// 	if err := svc.DeleteCompanyByID(context.Background(), id); err != nil {
// 		t.Fatalf("DeleteCompanyByID returned error: %v", err)
// 	}

// 	if len(pub.events) != 1 {
// 		t.Fatalf("expected 1 event, got %d", len(pub.events))
// 	}
// 	ev := pub.events[0]
// 	if ev.Operation != "company.deleted" {
// 		t.Fatalf("expected operation company.deleted, got %s", ev.Operation)
// 	}
// 	if ev.Company.ID != id {
// 		t.Fatalf("expected event company id %q, got %q", id, ev.Company.ID)
// 	}
// }

// func TestDeleteCompanyByID_NotFound_NoPublish(t *testing.T) {
// 	publisher := &stubPublisher{}
// 	repo := stubRepository{
// 		deleteFn: func(ctx context.Context, companyID string) error {
// 			return repoerrors.ErrNotFound
// 		},
// 	}

// 	svc := NewService(repo, publisher)

// 	err := svc.DeleteCompanyByID(context.Background(), "missing-id")
// 	if err == nil {
// 		t.Fatal("expected ErrNotFound but got nil")
// 	}

// 	if !errors.Is(err, ErrNotFound) {
// 		t.Fatalf("expected ErrNotFound, got %v", err)
// 	}

// 	if len(publisher.events) != 0 {
// 		t.Fatalf("expected no events when delete fails, got %d", len(publisher.events))
// 	}
// }

// // Patch tests
// func TestPatchCompanyByID_Success_Publishes_AndPassesArgs(t *testing.T) {
// 	const id = "company-123"
// 	partial := domain.PatchCompanyRequest{
// 		// adjust to your real fields
// 		Name: ptr("NewName"),
// 	}

// 	var seenPartial domain.PatchCompanyRequest
// 	var seenUUID string
// 	var seenMax int

// 	repo := stubRepository{
// 		patchFn: func(ctx context.Context, p domain.PatchCompanyRequest, uuid string, max int) error {
// 			seenPartial = p
// 			seenUUID = uuid
// 			seenMax = max
// 			return nil
// 		},
// 	}
// 	pub := &stubPublisher{}
// 	svc := NewService(repo, pub)

// 	if err := svc.PatchCompanyByID(context.Background(), partial, id); err != nil {
// 		t.Fatalf("PatchCompanyByID returned error: %v", err)
// 	}

// 	// Assert args to repo
// 	if seenUUID != id {
// 		t.Fatalf("uuid mismatch: got %q want %q", seenUUID, id)
// 	}
// 	if seenPartial.Name == nil || *seenPartial.Name != "NewName" {
// 		t.Fatalf("partial mismatch: got name=%v", seenPartial.Name)
// 	}
// 	if seenMax != 5 {
// 		t.Fatalf("expected maxNumOfFields=5, got %d", seenMax)
// 	}

// 	// Assert publish
// 	if len(pub.events) != 1 {
// 		t.Fatalf("expected 1 event, got %d", len(pub.events))
// 	}
// 	ev := pub.events[0]
// 	if ev.Operation != "company.patched" {
// 		t.Fatalf("expected operation company.patched, got %s", ev.Operation)
// 	}
// 	if ev.Company.ID != id {
// 		t.Fatalf("expected event company id %q, got %q", id, ev.Company.ID)
// 	}
// }

// func TestPatchCompanyByID_NotFound_NoPublish(t *testing.T) {
// 	repo := stubRepository{
// 		patchFn: func(ctx context.Context, _ domain.PatchCompanyRequest, _ string, _ int) error {
// 			return repository.ErrNotFound
// 		},
// 	}
// 	pub := &stubPublisher{}
// 	svc := NewService(repo, pub)

// 	err := svc.PatchCompanyByID(context.Background(), domain.PatchCompanyRequest{}, "missing-id")
// 	if err == nil || !errors.Is(err, ErrNotFound) {
// 		t.Fatalf("want ErrNotFound, got %v", err)
// 	}
// 	if len(pub.events) != 0 {
// 		t.Fatalf("should not publish when patch fails")
// 	}
// }

// func TestPatchCompanyByID_GenericRepoError_Bubbles_NoPublish(t *testing.T) {
// 	boom := errors.New("db blew up")
// 	repo := stubRepository{
// 		patchFn: func(ctx context.Context, _ domain.PatchCompanyRequest, _ string, _ int) error {
// 			return boom
// 		},
// 	}
// 	pub := &stubPublisher{}
// 	svc := NewService(repo, pub)

// 	err := svc.PatchCompanyByID(context.Background(), domain.PatchCompanyRequest{}, "any-id")
// 	if err == nil || !errors.Is(err, boom) {
// 		t.Fatalf("want original repo error %q, got %v", boom, err)
// 	}
// 	if len(pub.events) != 0 {
// 		t.Fatalf("should not publish when patch fails")
// 	}
// }

// Helper functions that allow to make assertions and report messages

func assertNoPublish(t *testing.T, pub *stubPublisher) {
	t.Helper()
	if got := len(pub.events); got != 0 {
		t.Fatalf("Then: expected no events to be published, got %d", got)
	}
}

func assertOneEvent(t *testing.T, pub *stubPublisher, op string) CompanyEvent {
	t.Helper()
	if got := len(pub.events); got != 1 {
		t.Fatalf("Then: expected 1 event, got %d", got)
	}
	ev := pub.events[0]
	if ev.Operation != op {
		t.Fatalf("Then: event operation: got %q, want %q", ev.Operation, op)
	}
	return ev
}

func assertDeepEqual[T any](t *testing.T, label string, got, want T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Then: %s mismatch\nGOT:  %+v\nWANT: %+v", label, got, want)
	}
}

func Given(t *testing.T, msg string, kv ...any) {
	t.Helper()
	if len(kv) > 0 {
		t.Logf("GIVEN: "+msg, kv...)
	} else {
		t.Logf("GIVEN: %s", msg)
	}
}

func When(t *testing.T, msg string, kv ...any) {
	t.Helper()
	if len(kv) > 0 {
		t.Logf("WHEN: "+msg, kv...)
	} else {
		t.Logf("WHEN: %s", msg)
	}
}

func Then(t *testing.T, msg string, kv ...any) {
	t.Helper()
	if len(kv) > 0 {
		t.Logf("THEN: "+msg, kv...)
	} else {
		t.Logf("THEN: %s", msg)
	}
}

func TestCreateCompany_Success_Publish(t *testing.T) {
	Given(t, "a valid company payload")

	expected := domain.Company{
		ID: "company-123", Name: "TechCorp",
		Description:       ptr("A software company"),
		AmountOfEmployees: 100, Registered: true, Type: domain.Corporations,
	}
	publisher := &stubPublisher{}
	repo := stubRepository{
		createFn: func(_ context.Context, company domain.Company) (domain.Company, error) {
			if company.Name != expected.Name {
				t.Fatalf("expected name %q but got %q", expected.Name, company.Name)
			}
			return expected, nil
		},
	}
	svc := NewService(repo, publisher)

	When(t, "CreateCompany is called with %q", expected.Name)
	got, err := svc.CreateCompany(context.Background(), expected)

	Then(t, "it returns the created company")
	if err != nil {
		t.Fatalf("CreateCompany returned unexpected error: %v", err)
	}
	assertDeepEqual(t, "created company", got, expected)

	Then(t, "it publishes 'company.created' with the same id and name")
	ev := assertOneEvent(t, publisher, "company.created")
	if ev.Company.ID != expected.ID || ev.Company.Name != expected.Name {
		t.Fatalf("unexpected event company payload: %+v", ev.Company)
	}
}

func TestCreateCompany_ValidationError_NoPublish(t *testing.T) {
	Given(t, "an invalid company (name too long)")

	repo := stubRepository{
		createFn: func(context.Context, domain.Company) (domain.Company, error) {
			t.Fatal("createFn should not be called when validation fails")
			return domain.Company{}, nil
		},
	}
	publisher := &stubPublisher{}
	svc := NewService(repo, publisher)
	longName := "ThisNameIsWayTooLong"
	company := domain.Company{
		ID: "id", Name: longName,
		Description: ptr("desc"), AmountOfEmployees: 10,
		Registered: true, Type: domain.Corporations,
	}

	When(t, "CreateCompany is called")
	_, err := svc.CreateCompany(context.Background(), company)

	Then(t, "it returns ErrValidationError and does not publish")
	if err == nil {
		t.Fatal("expected validation error but got nil")
	}
	if !errors.Is(err, ErrValidationError) {
		t.Fatalf("expected ErrValidationError, got %v", err)
	}
	assertNoPublish(t, publisher)
}

func TestCreateCompany_UniquenessViolation_NoPublish(t *testing.T) {
	Given(t, "a repo that returns a uniqueness violation")

	repoErr := repoerrors.ErrUniquenessViolation
	repo := stubRepository{
		createFn: func(context.Context, domain.Company) (domain.Company, error) {
			return domain.Company{}, repoErr
		},
	}
	publisher := &stubPublisher{}
	svc := NewService(repo, publisher)
	company := domain.Company{
		ID: "id", Name: "UniqueName",
		Description: ptr("desc"), AmountOfEmployees: 10,
		Registered: true, Type: domain.Corporations,
	}

	When(t, "CreateCompany is called")
	_, err := svc.CreateCompany(context.Background(), company)

	Then(t, "it maps to ErrUniquenessViolation and does not publish")
	if err == nil {
		t.Fatal("expected uniqueness violation error but got nil")
	}
	if !errors.Is(err, ErrUniquenessViolation) {
		t.Fatalf("expected ErrUniquenessViolation, got %v", err)
	}
	assertNoPublish(t, publisher)
}

func TestCreateCompany_RepoGenericError_Bubbles_NoPublish(t *testing.T) {
	Given(t, "a repo that returns a generic error")

	input := domain.Company{
		ID: "company-1", Name: "TechCorp", Type: domain.Corporations,
	}
	genericError := errors.New("db down")
	repo := stubRepository{
		createFn: func(context.Context, domain.Company) (domain.Company, error) {
			return domain.Company{}, genericError
		},
	}
	pub := &stubPublisher{}
	svc := NewService(repo, pub)

	When(t, "CreateCompany is called")
	_, err := svc.CreateCompany(context.Background(), input)

	Then(t, "it bubbles the original error and does not publish")
	if err == nil || !errors.Is(err, genericError) {
		t.Fatalf("want original repo error, got %v", err)
	}
	assertNoPublish(t, pub)
}

func TestGetCompanyByID_Success(t *testing.T) {
	Given(t, "an existing company id")

	want := domain.Company{
		ID: "company-123", Name: "TechCorp",
		Description:       ptr("A software company"),
		AmountOfEmployees: 100, Registered: true, Type: domain.Corporations,
	}
	repo := stubRepository{
		getFn: func(_ context.Context, id string) (domain.Company, error) {
			if id != want.ID {
				t.Fatalf("repo received id=%q, want %q", id, want.ID)
			}
			return want, nil
		},
	}
	svc := NewService(repo, &stubPublisher{})

	When(t, "GetCompanyByID is called")
	got, err := svc.GetCompanyByID(context.Background(), want.ID)

	Then(t, "it returns the company")
	if err != nil {
		t.Fatalf("GetCompanyByID returned error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetCompanyByID mismatch:\n got=%+v\nwant=%+v", got, want)
	}
}

func TestGetCompanyByID_NotFound(t *testing.T) {
	Given(t, "a missing company id")

	repo := stubRepository{
		getFn: func(context.Context, string) (domain.Company, error) {
			return domain.Company{}, repoerrors.ErrNotFound
		},
	}
	svc := NewService(repo, &stubPublisher{})

	When(t, "GetCompanyByID is called")
	_, err := svc.GetCompanyByID(context.Background(), "missing-id")

	Then(t, "it returns ErrNotFound")
	if err == nil {
		t.Fatal("expected ErrNotFound but got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetCompanyByID_GenericRepoError_BubblesUp(t *testing.T) {
	Given(t, "a repo that fails generically on GetCompanyByID")

	genericError := errors.New("db connection failed")
	repo := stubRepository{
		getFn: func(context.Context, string) (domain.Company, error) {
			return domain.Company{}, genericError
		},
	}
	svc := NewService(repo, &stubPublisher{})

	When(t, "GetCompanyByID is called")
	_, err := svc.GetCompanyByID(context.Background(), "any-id")

	Then(t, "it bubbles the original error")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, genericError) {
		t.Fatalf("want original repo error %q, got %v", genericError, err)
	}
}

func TestDeleteCompanyByID_Success_Publish(t *testing.T) {
	Given(t, "an existing company id to delete")

	const id = "company-123"
	repo := stubRepository{
		deleteFn: func(_ context.Context, got string) error {
			if got != id {
				t.Fatalf("repo received id %q, want %q", got, id)
			}
			return nil
		},
	}
	pub := &stubPublisher{}
	svc := NewService(repo, pub)

	When(t, "DeleteCompanyByID is called")
	if err := svc.DeleteCompanyByID(context.Background(), id); err != nil {
		t.Fatalf("DeleteCompanyByID returned error: %v", err)
	}

	Then(t, "it publishes 'company.deleted' with the id")
	ev := assertOneEvent(t, pub, "company.deleted")
	if ev.Company.ID != id {
		t.Fatalf("expected event company id %q, got %q", id, ev.Company.ID)
	}
}

func TestDeleteCompanyByID_NotFound_NoPublish(t *testing.T) {
	Given(t, "a missing company id for deletion")

	publisher := &stubPublisher{}
	repo := stubRepository{
		deleteFn: func(context.Context, string) error { return repoerrors.ErrNotFound },
	}
	svc := NewService(repo, publisher)

	When(t, "DeleteCompanyByID is called")
	err := svc.DeleteCompanyByID(context.Background(), "missing-id")

	Then(t, "it returns ErrNotFound and does not publish")
	if err == nil {
		t.Fatal("expected ErrNotFound but got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	assertNoPublish(t, publisher)
}

func TestPatchCompanyByID_NoFieldsProvided_ReturnsValidationError_NoPublish(t *testing.T) {
	Given(t, "a patch payload with no fields")

	repo := stubRepository{
		patchFn: func(context.Context, domain.PatchCompanyRequest, string, int) error {
			t.Fatal("patchFn should not be called when validation fails")
			return nil
		},
	}
	pub := &stubPublisher{}
	svc := NewService(repo, pub)

	When(t, "PatchCompanyByID is called")
	err := svc.PatchCompanyByID(context.Background(), domain.PatchCompanyRequest{}, "company-123")

	Then(t, "it returns ErrValidationError and does not publish")
	if err == nil {
		t.Fatal("expected validation error but got nil")
	}
	if !errors.Is(err, ErrValidationError) {
		t.Fatalf("expected ErrValidationError, got %v", err)
	}
	assertNoPublish(t, pub)
}

func TestPatchCompanyByID_InvalidName_ReturnsValidationError_NoPublish(t *testing.T) {
	Given(t, "a patch payload with an invalid name")

	repo := stubRepository{
		patchFn: func(context.Context, domain.PatchCompanyRequest, string, int) error {
			t.Fatal("patchFn should not be called when validation fails")
			return nil
		},
	}
	pub := &stubPublisher{}
	svc := NewService(repo, pub)

	partial := domain.PatchCompanyRequest{Name: ptr("1234567890123456")}

	When(t, "PatchCompanyByID is called")
	err := svc.PatchCompanyByID(context.Background(), partial, "company-123")

	Then(t, "it returns ErrValidationError and does not publish")
	if err == nil {
		t.Fatal("expected validation error but got nil")
	}
	if !errors.Is(err, ErrValidationError) {
		t.Fatalf("expected ErrValidationError, got %v", err)
	}
	assertNoPublish(t, pub)
}

func TestPatchCompanyByID_Success_Publishes_AndPassesArgs(t *testing.T) {
	Given(t, "a valid patch payload and existing id")

	const id = "company-123"
	partial := domain.PatchCompanyRequest{Name: ptr("NewName")}

	var seenPartial domain.PatchCompanyRequest
	var seenUUID string
	var seenMax int

	repo := stubRepository{
		patchFn: func(_ context.Context, p domain.PatchCompanyRequest, uuid string, max int) error {
			seenPartial = p
			seenUUID = uuid
			seenMax = max
			return nil
		},
	}
	pub := &stubPublisher{}
	svc := NewService(repo, pub)

	When(t, "PatchCompanyByID is called")
	if err := svc.PatchCompanyByID(context.Background(), partial, id); err != nil {
		t.Fatalf("PatchCompanyByID returned error: %v", err)
	}

	Then(t, "it passes args to repo and publishes 'company.patched'")
	if seenUUID != id {
		t.Fatalf("uuid mismatch: got %q want %q", seenUUID, id)
	}
	if seenPartial.Name == nil || *seenPartial.Name != "NewName" {
		t.Fatalf("partial mismatch: got name=%v", seenPartial.Name)
	}
	if seenMax != 5 {
		t.Fatalf("expected maxNumOfFields=5, got %d", seenMax)
	}
	ev := assertOneEvent(t, pub, "company.patched")
	if ev.Company.ID != id {
		t.Fatalf("expected event company id %q, got %q", id, ev.Company.ID)
	}
}

func TestPatchCompanyByID_NotFound_NoPublish(t *testing.T) {
	Given(t, "a repo that returns not found on patch")

	repo := stubRepository{
		patchFn: func(context.Context, domain.PatchCompanyRequest, string, int) error {
			return repoerrors.ErrNotFound
		},
	}
	pub := &stubPublisher{}
	svc := NewService(repo, pub)

	When(t, "PatchCompanyByID is called")
	err := svc.PatchCompanyByID(context.Background(), domain.PatchCompanyRequest{Name: ptr("ExistingName")}, "missing-id")

	Then(t, "it returns ErrNotFound and does not publish")
	if err == nil || !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	assertNoPublish(t, pub)
}

func TestPatchCompanyByID_GenericRepoError_Bubbles_NoPublish(t *testing.T) {
	Given(t, "a repo that fails with a generic error on patch")

	boom := errors.New("db blew up")
	repo := stubRepository{
		patchFn: func(context.Context, domain.PatchCompanyRequest, string, int) error {
			return boom
		},
	}
	pub := &stubPublisher{}
	svc := NewService(repo, pub)

	When(t, "PatchCompanyByID is called")
	err := svc.PatchCompanyByID(context.Background(), domain.PatchCompanyRequest{Name: ptr("ExistingName")}, "any-id")

	Then(t, "it bubbles the original error and does not publish")
	if err == nil || !errors.Is(err, boom) {
		t.Fatalf("want original repo error %q, got %v", boom, err)
	}
	assertNoPublish(t, pub)
}
