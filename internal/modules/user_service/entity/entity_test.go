package entity_test

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	internalEntity "github.com/k1v4/drip_mate/internal/entity"
	userEntity "github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validate = validator.New()

func TestUser_JSON_PasswordHidden(t *testing.T) {
	user := userEntity.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Password: "supersecret",
		Username: "kirill",
		Name:     "Kirill",
		Surname:  "Ivanov",
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	assert.NotContains(t, string(data), "supersecret")
	assert.NotContains(t, string(data), "password")

	assert.Contains(t, string(data), "user@example.com")
	assert.Contains(t, string(data), "kirill")
}

func TestUser_JSON_NullableSlices(t *testing.T) {
	user := userEntity.User{
		ID:    uuid.New(),
		Email: "user@example.com",
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Nil(t, raw["music"])
	assert.Nil(t, raw["styles"])
	assert.Nil(t, raw["colors"])
	assert.Nil(t, raw["outfits"])
}

func TestUser_JSON_FilledSlices(t *testing.T) {
	user := userEntity.User{
		ID:     uuid.New(),
		Music:  []string{"jazz", "rock"},
		Styles: []string{"casual"},
		Colors: []string{"black", "white"},
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))

	music, ok := raw["music"].([]any)
	require.True(t, ok)
	assert.Len(t, music, 2)
	assert.Equal(t, "jazz", music[0])
}

func TestUser_MarshalBinary(t *testing.T) {
	original := &userEntity.User{
		ID:          uuid.New(),
		Email:       "user@example.com",
		Password:    "secret",
		Username:    "kirill",
		Name:        "Kirill",
		Surname:     "Ivanov",
		City:        "Perm",
		AccessID:    1,
		AccessLevel: "user",
		Music:       []string{"jazz"},
		Styles:      []string{"casual"},
		Colors:      []string{"black"},
	}

	data, err := original.MarshalBinary()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored := &userEntity.User{}
	err = restored.UnmarshalBinary(data)
	require.NoError(t, err)

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Email, restored.Email)
	assert.Equal(t, original.Username, restored.Username)
	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Surname, restored.Surname)
	assert.Equal(t, original.City, restored.City)
	assert.Equal(t, original.AccessID, restored.AccessID)
	assert.Equal(t, original.AccessLevel, restored.AccessLevel)
	assert.Equal(t, original.Music, restored.Music)
	assert.Equal(t, original.Styles, restored.Styles)
	assert.Equal(t, original.Colors, restored.Colors)
}

func TestUser_MarshalBinary_PasswordExcluded(t *testing.T) {
	user := &userEntity.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Password: "supersecret",
	}

	data, err := user.MarshalBinary()
	require.NoError(t, err)
	assert.NotContains(t, string(data), "supersecret")
}

func TestUser_UnmarshalBinary_InvalidData(t *testing.T) {
	user := &userEntity.User{}
	err := user.UnmarshalBinary([]byte("not valid json{{{"))
	assert.Error(t, err)
}

func TestUser_UnmarshalBinary_EmptyData(t *testing.T) {
	user := &userEntity.User{}
	err := user.UnmarshalBinary([]byte("{}"))
	require.NoError(t, err)

	assert.Equal(t, uuid.Nil, user.ID)
	assert.Empty(t, user.Email)
}

func TestRegisterRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     userEntity.RegisterRequest
		wantErr bool
	}{
		{
			name:    "valid",
			req:     userEntity.RegisterRequest{Email: "user@example.com", Password: "strongpass1"},
			wantErr: false,
		},
		{
			name:    "missing email",
			req:     userEntity.RegisterRequest{Password: "strongpass1"},
			wantErr: true,
		},
		{
			name:    "invalid email format",
			req:     userEntity.RegisterRequest{Email: "not-an-email", Password: "strongpass1"},
			wantErr: true,
		},
		{
			name:    "missing password",
			req:     userEntity.RegisterRequest{Email: "user@example.com"},
			wantErr: true,
		},
		{
			name:    "password too short",
			req:     userEntity.RegisterRequest{Email: "user@example.com", Password: "short"},
			wantErr: true,
		},
		{
			name:    "password exactly 10 chars",
			req:     userEntity.RegisterRequest{Email: "user@example.com", Password: "exactly10c"},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Struct(tc.req)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdatePass_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     userEntity.UpdatePass
		wantErr bool
	}{
		{
			name:    "valid",
			req:     userEntity.UpdatePass{CurrPassword: "oldpassword", NewPassword: "newpassword1"},
			wantErr: false,
		},
		{
			name:    "missing curr_password",
			req:     userEntity.UpdatePass{NewPassword: "newpassword1"},
			wantErr: true,
		},
		{
			name:    "missing new_password",
			req:     userEntity.UpdatePass{CurrPassword: "oldpassword"},
			wantErr: true,
		},
		{
			name:    "new_password too short",
			req:     userEntity.UpdatePass{CurrPassword: "oldpassword", NewPassword: "short"},
			wantErr: true,
		},
		{
			name:    "new_password exactly 10 chars",
			req:     userEntity.UpdatePass{CurrPassword: "oldpassword", NewPassword: "exactly10c"},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Struct(tc.req)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSaveOutfitRequest_Validate(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name    string
		req     userEntity.SaveOutfitRequest
		wantErr bool
	}{
		{
			name:    "valid",
			req:     userEntity.SaveOutfitRequest{Name: "Summer", CatalogItemIDs: []uuid.UUID{validID}},
			wantErr: false,
		},
		{
			name:    "missing name",
			req:     userEntity.SaveOutfitRequest{CatalogItemIDs: []uuid.UUID{validID}},
			wantErr: true,
		},
		{
			name:    "name too long — 256 chars",
			req:     userEntity.SaveOutfitRequest{Name: string(make([]byte, 256)), CatalogItemIDs: []uuid.UUID{validID}},
			wantErr: true,
		},
		{
			name:    "name exactly 255 chars",
			req:     userEntity.SaveOutfitRequest{Name: string(make([]byte, 255)), CatalogItemIDs: []uuid.UUID{validID}},
			wantErr: false,
		},
		{
			name:    "missing catalog_item_ids",
			req:     userEntity.SaveOutfitRequest{Name: "Summer"},
			wantErr: true,
		},
		{
			name:    "empty catalog_item_ids",
			req:     userEntity.SaveOutfitRequest{Name: "Summer", CatalogItemIDs: []uuid.UUID{}},
			wantErr: true,
		},
		{
			name:    "nil uuid in catalog_item_ids",
			req:     userEntity.SaveOutfitRequest{Name: "Summer", CatalogItemIDs: []uuid.UUID{uuid.Nil}},
			wantErr: true, // dive,required отклоняет нулевой UUID
		},
		{
			name:    "with optional LogID",
			req:     userEntity.SaveOutfitRequest{Name: "Summer", CatalogItemIDs: []uuid.UUID{validID}, LogID: 42},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Struct(tc.req)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOutfit_JSON_RoundTrip(t *testing.T) {
	original := userEntity.Outfit{
		ID:   uuid.New(),
		Name: "Winter outfit",
		Items: []userEntity.OutfitItem{
			{
				ID:       uuid.New(),
				Name:     "Jacket",
				ImageURL: "https://example.com/jacket.jpg",
				Material: "wool",
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored userEntity.Outfit
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Name, restored.Name)
	assert.Len(t, restored.Items, 1)
	assert.Equal(t, original.Items[0].ID, restored.Items[0].ID)
	assert.Equal(t, original.Items[0].Name, restored.Items[0].Name)
	assert.Equal(t, original.Items[0].ImageURL, restored.Items[0].ImageURL)
	assert.Equal(t, original.Items[0].Material, restored.Items[0].Material)
}

func TestOutfit_JSON_EmptyItems(t *testing.T) {
	outfit := userEntity.Outfit{
		ID:    uuid.New(),
		Name:  "Empty outfit",
		Items: []userEntity.OutfitItem{},
	}

	data, err := json.Marshal(outfit)
	require.NoError(t, err)

	var restored userEntity.Outfit
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.NotNil(t, restored.Items)
	assert.Empty(t, restored.Items)
}

func TestUpdateContext_JSON_NilPointers(t *testing.T) {
	ctx := userEntity.UpdateContext{
		ID:   uuid.New(),
		City: "Moscow",
		// Styles, Colors, Music — nil
	}

	data, err := json.Marshal(ctx)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Nil(t, raw["styles"])
	assert.Nil(t, raw["colors"])
	assert.Nil(t, raw["music"])
	assert.Equal(t, "Moscow", raw["city"])
}

func TestUpdateContext_JSON_FilledPointers(t *testing.T) {
	styles := []int{1, 2, 3}
	colors := []int{4, 5}
	music := []int{6}

	ctx := userEntity.UpdateContext{
		ID:     uuid.New(),
		City:   "Perm",
		Styles: &styles,
		Colors: &colors,
		Music:  &music,
	}

	data, err := json.Marshal(ctx)
	require.NoError(t, err)

	var restored userEntity.UpdateContext
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NotNil(t, restored.Styles)
	assert.Equal(t, styles, *restored.Styles)

	require.NotNil(t, restored.Colors)
	assert.Equal(t, colors, *restored.Colors)

	require.NotNil(t, restored.Music)
	assert.Equal(t, music, *restored.Music)
}

func TestLoginResponse_JSON(t *testing.T) {
	resp := userEntity.LoginResponse{
		AccessId: internalEntity.Role(2),
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))

	accessID, ok := raw["access_id"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(2), accessID)
}

func TestDeleteUserResponse_JSON(t *testing.T) {
	tests := []struct {
		name           string
		isSuccessfully bool
	}{
		{name: "success true", isSuccessfully: true},
		{name: "success false", isSuccessfully: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := userEntity.DeleteUserResponse{IsSuccessfully: tc.isSuccessfully}

			data, err := json.Marshal(resp)
			require.NoError(t, err)

			var restored userEntity.DeleteUserResponse
			require.NoError(t, json.Unmarshal(data, &restored))

			assert.Equal(t, tc.isSuccessfully, restored.IsSuccessfully)
		})
	}
}
