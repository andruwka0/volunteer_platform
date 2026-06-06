package platform

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// newTestApp создаёт изолированное приложение с временным JSONStore для тестов.
func newTestApp(t *testing.T) *App {
	t.Helper()
	db, err := OpenDB(Config{DatabaseURL: t.TempDir() + "/db.json", SecretKey: "test-secret", SessionCookieName: "test_session", Addr: ":0"})
	if err != nil {
		t.Fatal(err)
	}
	return NewApp(db, Config{SecretKey: "test-secret", SessionCookieName: "test_session", Addr: ":0"})
}

// csrfFromBody извлекает hidden CSRF token из HTML формы.
func csrfFromBody(t *testing.T, body string) string {
	t.Helper()
	re := regexp.MustCompile(`name="csrf_token" value="([^"]+)"`)
	m := re.FindStringSubmatch(body)
	if len(m) != 2 {
		t.Fatalf("csrf token not found in body: %s", body)
	}
	return m[1]
}

// loginAs выполняет GET/POST login flow и возвращает session cookie.
func loginAs(t *testing.T, app *App, username, password string) *http.Cookie {
	t.Helper()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	app.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /login status = %d", rr.Code)
	}
	csrf := csrfFromBody(t, rr.Body.String())
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login page did not set a session cookie")
	}
	// Use the final cookie the way a browser would if multiple Set-Cookie
	// headers are present for the same name.
	sessionCookie := cookies[len(cookies)-1]
	form := url.Values{"username": {username}, "password": {password}, "csrf_token": {csrf}}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	app.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("POST /login status = %d, body=%s", rr.Code, rr.Body.String())
	}
	return sessionCookie
}

// TestLoginPageUsesOneSessionForCSRFAndCookie описывает назначение одноимённой функции.
func TestLoginPageUsesOneSessionForCSRFAndCookie(t *testing.T) {
	app := newTestApp(t)
	_, err := CreateUser(app.DB, "leader", "Leader", "Password123", "leader", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/login", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /login status = %d", rr.Code)
	}
	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("GET /login set %d cookies, want exactly 1; cookies=%v", len(cookies), cookies)
	}
	csrf := csrfFromBody(t, rr.Body.String())
	form := url.Values{"username": {"leader"}, "password": {"Password123"}, "csrf_token": {csrf}}
	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[len(cookies)-1])
	app.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("POST /login with browser-final cookie status = %d, body=%s", rr.Code, rr.Body.String())
	}
}

// TestLoginWithValidCredentialsDoesNotRequireFreshCSRF описывает назначение одноимённой функции.
func TestLoginWithValidCredentialsDoesNotRequireFreshCSRF(t *testing.T) {
	app := newTestApp(t)
	_, err := CreateUser(app.DB, "leader", "Leader", "Password123", "leader", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{"username": {"leader"}, "password": {"Password123"}, "csrf_token": {"stale-or-missing-token"}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("POST /login with stale CSRF status = %d, body=%s", rr.Code, rr.Body.String())
	}
	if len(rr.Result().Cookies()) == 0 {
		t.Fatal("successful login did not set a session cookie")
	}
}

// TestAuthenticatedActionsStillRequireCSRF описывает назначение одноимённой функции.
func TestAuthenticatedActionsStillRequireCSRF(t *testing.T) {
	app := newTestApp(t)
	_, err := CreateUser(app.DB, "leader", "Leader", "Password123", "leader", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	cookie := loginAs(t, app, "leader", "Password123")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(cookie)
	app.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("POST /logout without CSRF status = %d, want 403", rr.Code)
	}
}

// TestHealthLoginAndDashboard описывает назначение одноимённой функции.
func TestHealthLoginAndDashboard(t *testing.T) {
	app := newTestApp(t)
	_, err := CreateUser(app.DB, "leader", "Leader", "Password123", "leader", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rr.Code != http.StatusOK || strings.TrimSpace(rr.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("health response = status %d body %q", rr.Code, rr.Body.String())
	}

	cookie := loginAs(t, app, "leader", "Password123")
	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	app.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "Привет, Leader") {
		t.Fatalf("dashboard status=%d body=%s", rr.Code, rr.Body.String())
	}
}

// TestAdminEventEditUsesAdminPathID описывает назначение одноимённой функции.
func TestAdminEventEditUsesAdminPathID(t *testing.T) {
	app := newTestApp(t)
	leader, err := CreateUser(app.DB, "leader", "Leader", "Password123", "leader", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	now := Now()
	e := Event{ID: app.DB.NextID("events"), Title: "Original", ShortDescription: "Short", FullDescription: "Full", StartDate: now.Add(24 * time.Hour), EndDate: now.Add(48 * time.Hour), RatingDeadline: now.Add(48 * time.Hour), CreatedByID: leader.ID, Status: "soon", CreatedAt: now, UpdatedAt: now}
	app.DB.Events = append(app.DB.Events, e)
	app.DB.Save()
	cookie := loginAs(t, app, "leader", "Password123")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/events/1/edit", nil)
	req.AddCookie(cookie)
	app.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "Original") {
		t.Fatalf("GET edit status=%d body=%s", rr.Code, rr.Body.String())
	}
	csrf := csrfFromBody(t, rr.Body.String())
	form := url.Values{
		"csrf_token":        {csrf},
		"title":             {"Updated"},
		"short_description": {"Short"},
		"full_description":  {"Full"},
		"start_date":        {now.Add(72 * time.Hour).Format("2006-01-02")},
		"end_date":          {now.Add(96 * time.Hour).Format("2006-01-02")},
		"sp_points":         {"15"},
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/events/1/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	app.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		body, _ := io.ReadAll(rr.Result().Body)
		t.Fatalf("POST edit status=%d body=%s", rr.Code, body)
	}
	if got := app.DB.Event(1).Title; got != "Updated" {
		t.Fatalf("event title = %q, want Updated", got)
	}
}

// TestRestoredProfileMerchRulesAndEventFormUI описывает назначение одноимённой функции.
func TestRestoredProfileMerchRulesAndEventFormUI(t *testing.T) {
	app := newTestApp(t)
	leader, err := CreateUser(app.DB, "leader", "Leader", "Password123", "leader", "", "@leader", "", "")
	if err != nil {
		t.Fatal(err)
	}
	volunteer, err := CreateUser(app.DB, "volunteer", "Volunteer", "Password123", "junior_volunteer", "/static/img/placeholder.svg", "@vol", "", "")
	if err != nil {
		t.Fatal(err)
	}
	volunteer.SkillPoints = 12
	now := Now()
	e := Event{ID: app.DB.NextID("events"), Title: "Role event", ShortDescription: "Short", FullDescription: "Full", StartDate: now.Add(24 * time.Hour), EndDate: now.Add(48 * time.Hour), RatingDeadline: now.Add(48 * time.Hour), CreatedByID: leader.ID, Status: "soon", CreatedAt: now, UpdatedAt: now}
	capValue := 2
	app.DB.Events = append(app.DB.Events, e)
	app.DB.EventRoles = append(app.DB.EventRoles, EventRole{ID: app.DB.NextID("event_roles"), EventID: e.ID, Title: "Фото", Description: "Снимать мероприятие", Capacity: &capValue, IsActive: true})
	app.DB.Save()

	leaderCookie := loginAs(t, app, "leader", "Password123")
	for _, tc := range []struct {
		path string
		want []string
	}{
		{"/admin/events/new", []string{"name=\"start_date\"", "name=\"end_date\"", "name=\"event_image_file\"", "name=\"event_cropped_data\"", "data-crop-kind=\"event\"", "data-event-roles-builder", "name=\"role_titles\"", "data-multi-select"}},
		{"/admin/events/1/edit", []string{"Role event", "name=\"start_date\"", "event-image-preview", "Фото", "Снимать мероприятие"}},
		{"/rules", []string{"Ваши возможности", "Усиление · 3+ баллов", "Вовлечение · 10+ баллов", "Сотворчество · 20+ баллов"}},
		{"/admin/roles", []string{"leader", "organizer", "senior_volunteer", "middle_volunteer", "junior_volunteer"}},
	} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		req.AddCookie(leaderCookie)
		app.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d", tc.path, rr.Code)
		}
		body := rr.Body.String()
		for _, want := range tc.want {
			if !strings.Contains(body, want) {
				t.Fatalf("GET %s missing %q in body: %s", tc.path, want, body)
			}
		}
	}

	volunteerCookie := loginAs(t, app, "volunteer", "Password123")
	for _, tc := range []struct {
		path string
		want []string
	}{
		{"/profile", []string{"profile-hero", "profile-avatar", "Мои<br>Skill Points", "Персонализация профиля", "avatar-cropped-data", "data-crop-kind=\"avatar\""}},
		{"/leaderboard", []string{"Мерч за Skill Points", "Твои SP", "merch-progress", "Бутылка / термокружка"}},
		{"/events/1", []string{"Фото", "Снимать мероприятие", "desired_role_ids"}},
	} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		req.AddCookie(volunteerCookie)
		app.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d", tc.path, rr.Code)
		}
		body := rr.Body.String()
		for _, want := range tc.want {
			if !strings.Contains(body, want) {
				t.Fatalf("GET %s missing %q in body: %s", tc.path, want, body)
			}
		}
	}
}

// postAdminEventEditMultipart отправляет multipart admin-edit форму с CSRF token.
func postAdminEventEditMultipart(t *testing.T, app *App, cookie *http.Cookie, eventID int, fields map[string]string, fileField bool) *httptest.ResponseRecorder {
	t.Helper()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/events/"+itoa(eventID)+"/edit", nil)
	req.AddCookie(cookie)
	app.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET edit for event %d status=%d body=%s", eventID, rr.Code, rr.Body.String())
	}
	fields["csrf_token"] = csrfFromBody(t, rr.Body.String())

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if fileField {
		part, err := writer.CreateFormFile("event_image_file", "new-cover.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte("not a real png, but enough for upload persistence")); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/events/"+itoa(eventID)+"/edit", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	app.Routes().ServeHTTP(rr, req)
	return rr
}

// itoa сокращает strconv.Itoa для сборки тестовых URL.
func itoa(v int) string { return strconv.Itoa(v) }

// TestPastAndOngoingEventCoverImageCanBeUpdated описывает назначение одноимённой функции.
func TestPastAndOngoingEventCoverImageCanBeUpdated(t *testing.T) {
	app := newTestApp(t)
	leader, err := CreateUser(app.DB, "leader", "Leader", "Password123", "leader", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	now := Now()
	events := []Event{
		{ID: app.DB.NextID("events"), Title: "Past", ShortDescription: "Short", FullDescription: "Full", StartDate: now.Add(-72 * time.Hour), EndDate: now.Add(-48 * time.Hour), RatingDeadline: now.Add(-48 * time.Hour), CoverImageURL: "/static/img/old-past.svg", CreatedByID: leader.ID, Status: "closed", CreatedAt: now, UpdatedAt: now},
		{ID: app.DB.NextID("events"), Title: "Ongoing", ShortDescription: "Short", FullDescription: "Full", StartDate: now.Add(-1 * time.Hour), EndDate: now.Add(1 * time.Hour), RatingDeadline: now.Add(1 * time.Hour), CoverImageURL: "/static/img/old-ongoing.svg", CreatedByID: leader.ID, Status: "live", CreatedAt: now, UpdatedAt: now},
	}
	app.DB.Events = append(app.DB.Events, events...)
	app.DB.Save()
	cookie := loginAs(t, app, "leader", "Password123")

	croppedPNG := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII="
	cases := []struct {
		name      string
		event     Event
		cropped   string
		fileField bool
	}{
		{name: "past cropped image", event: events[0], cropped: croppedPNG},
		{name: "ongoing file upload", event: events[1], fileField: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fields := map[string]string{
				"title":              tc.event.Title,
				"short_description":  tc.event.ShortDescription,
				"full_description":   tc.event.FullDescription,
				"start_date":         tc.event.StartDate.Format("2006-01-02"),
				"end_date":           tc.event.EndDate.Format("2006-01-02"),
				"sp_points":          "10",
				"no_roles":           "on",
				"event_cropped_data": tc.cropped,
			}
			rr := postAdminEventEditMultipart(t, app, cookie, tc.event.ID, fields, tc.fileField)
			if rr.Code != http.StatusFound {
				t.Fatalf("POST edit status=%d body=%s", rr.Code, rr.Body.String())
			}
			updated := app.DB.Event(tc.event.ID)
			if updated.CoverImageURL == tc.event.CoverImageURL {
				t.Fatalf("cover image stayed unchanged: %q", updated.CoverImageURL)
			}
			if !strings.HasPrefix(updated.CoverImageURL, "/static/uploads/event_") {
				t.Fatalf("cover image = %q, want uploaded event path", updated.CoverImageURL)
			}
			t.Cleanup(func() {
				_ = os.Remove(filepath.Join(findProjectRoot(), "app", "static", "uploads", filepath.Base(updated.CoverImageURL)))
			})
		})
	}
}

// TestCropperUIUsesOnlyZoomControlAndDragOffsets описывает назначение одноимённой функции.
func TestCropperUIUsesOnlyZoomControlAndDragOffsets(t *testing.T) {
	js, err := os.ReadFile(filepath.Join(findProjectRoot(), "app", "static", "js", "main.js"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(js)
	for _, forbidden := range []string{"data-crop-x", "data-crop-y", "Смещение по X", "Смещение по Y"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("cropper JS still contains removed coordinate control %q", forbidden)
		}
	}
	for _, want := range []string{"data-crop-zoom", "state.offsetX", "state.offsetY", "Можно двигать картинку мышкой/пальцем и менять масштаб"} {
		if !strings.Contains(content, want) {
			t.Fatalf("cropper JS missing %q", want)
		}
	}
}
