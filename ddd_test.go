package ddd_test

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	ddd "github.com/vklap/go-ddd"
	"os"
	"testing"
)

func cleanup() {
	_ = ddd.CleanupDbTables()
}

func TestMain(m *testing.M) {
	ddd.Hostname = "localhost"
	ddd.Password = "IWillLoveGo4Ever!"
	ddd.Database = "go"
	ddd.Username = "go_admin"
	err := ddd.CleanupDbTables()
	if err != nil {
		fmt.Println("failed to cleanup DB tables:", err)
		os.Exit(1)
	}
	exitVal := m.Run()
	os.Exit(exitVal)
}

func Test_AddUser(t *testing.T) {
	t.Cleanup(cleanup)
	d := ddd.UserData{
		Username:    "vklap",
		Name:        "Victor",
		Surname:     "Klapholz",
		Description: "Go programmer",
	}

	err := ddd.AddUser(&d)

	if err != nil {
		t.Error("user was not created, got:", err)
	}
	if d.ID <= 0 {
		t.Error("user id was not filled correctly, got:", d.ID)
	}
}

func Test_GetUserByUsername(t *testing.T) {
	t.Cleanup(cleanup)
	want := ddd.UserData{
		Username:    "vklap",
		Name:        "Victor",
		Surname:     "Klapholz",
		Description: "Go programmer",
	}
	err := ddd.AddUser(&want)
	if err != nil {
		t.Error("user was not created, got:", err)
	}

	got, err := ddd.GetUserByUsername("VKLAP")

	if err != nil {
		t.Error("failed to get user, got:", err)
	}
	if diff := cmp.Diff(want, *got); diff != "" {
		t.Errorf("GetUserByUsername() mismatch (-got +got):\n%s", diff)
	}
}
