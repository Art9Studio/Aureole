package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"aureole/internal/plugins/storage"
	"aureole/internal/plugins/storage/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createConnSess(t *testing.T) types.Storage {
	conf := &configs.Storage{
		Type: "postgresql",
		Name: "",
		Config: configs.RawConfig{
			"url": "postgresql://root:password@localhost:5432/aureole?sslmode=disable&search_path=public",
		},
	}

	s, err := storage.New(conf)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}

	if err := s.Init(); err != nil {
		t.Fatalf("open connection by url: %v", err)
	}

	return s
}

func Test_Session_IsCollExists(t *testing.T) {
	usersSess := createConnSess(t)
	defer usersSess.Close()

	res, err := usersSess.IsCollExists(collections.Spec{Name: "users", Pk: "id"})
	assert.NoError(t, err)
	assert.Equal(t, res, true)

	res, err = usersSess.IsCollExists(collections.Spec{Name: "users", Pk: "id"})
	assert.NoError(t, err)
	assert.Equal(t, res, false)
}

/*
func Test_Session_CreateIdentitytColl(t *testing.T) {
	usersSess := createConnSess(t)
	defer usersSess.Close()

	i := &identity.Identity{
		Id: identity.Trait{
			Enabled:  true,
			Unique:   true,
			Required: true,
			Internal: true,
		},
		Username: identity.Trait{
			Enabled:  true,
			Unique:   false,
			Required: false,
			Internal: false,
		},
		Phone: identity.Trait{
			Enabled:  false,
			Unique:   true,
			Required: false,
			Internal: false,
		},
		Email: identity.Trait{
			Enabled:  true,
			Unique:   true,
			Required: true,
			Internal: false,
		},
	}

	i.Collection = &collections.Collection{
		Spec: collections.Spec{
			Name:      "users",
			Pk:        "id",
			FieldsMap: map[string]string{"username": "username", "phone": "phone", "email": "email"},
		},
	}
	err := usersSess.CreateIdentityColl(i)
	assert.Contains(t, err.Error(), "already exists")
	/*
		err = usersSess.CreateIdentityColl(collections.Spec{
			Name:      "other",
			Pk:        "id",
			FieldsMap: map[string]string{"identity": "username", "password": "password"},
		})
		assert.NoError(t, err)

		err = usersSess.CreateIdentityColl(collections.Spec{
			Name:      "); drop table other; --",
			Pk:        "id",
			FieldsMap: map[string]string{"identity": "username", "password": "password"},
		})
		assert.NoError(t, err)

		isOtherExist, err := usersSess.IsCollExists(collections.Spec{Name: "other", Pk: "id"})
		assert.NoError(t, err)
		assert.True(t, isOtherExist)

		isDropExist, err := usersSess.IsCollExists(collections.Spec{Name: "); drop table other; --", Pk: "id"})
		assert.NoError(t, err)
		assert.True(t, isDropExist)

}

func Test_Session_InsertIdentity(t *testing.T) {
	usersSess := createConnSess(t)
	defer usersSess.Close()

	res, err := usersSess.InsertIdentity(
		collTypes.Spec{
			Name:      "users",
			Pk:        "id",
			FieldsMap: map[string]string{"identity": "username", "password": "password"},
		},
		types.IdentityData{Identity: "hello", Phone: "secret"},
	)
	assert.NoError(t, err)
	fmt.Printf("new id: %v\n", res)
}
*/
