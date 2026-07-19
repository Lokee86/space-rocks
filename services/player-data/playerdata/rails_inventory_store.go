package playerdata

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func (s *RailsStore) LoadHangarInventory(identity protocol.PlayerDataIdentity) (protocol.HangarInventory, bool, error) {
	if identity.IdentityKind != IdentityKindAuthenticatedAccount {
		return protocol.HangarInventory{}, false, NewClassifiedFailure(FailureClassInvalidResponse, errors.New("identity_kind must be authenticated_account"))
	}
	if identity.AccountID == "" {
		return protocol.HangarInventory{}, false, NewClassifiedFailure(FailureClassInvalidResponse, errors.New("account_id is required"))
	}
	if s.internalToken == "" {
		return protocol.HangarInventory{}, false, NewClassifiedFailure(FailureClassAuthentication, errors.New("internal token is required"))
	}

	request, err := s.newJSONRequest(http.MethodPost, "/api/internal/player-data/inventory/load", struct {
		AccountID string `json:"account_id"`
	}{AccountID: identity.AccountID})
	if err != nil {
		return protocol.HangarInventory{}, false, err
	}
	response, err := s.client().Do(request)
	if err != nil {
		return protocol.HangarInventory{}, false, NewClassifiedFailure(FailureClassUpstreamUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return protocol.HangarInventory{}, false, nil
	}
	if response.StatusCode != http.StatusOK {
		return protocol.HangarInventory{}, false, NewClassifiedFailure(FailureClassUnexpectedStatus, errors.New("unexpected status"))
	}

	var decoded struct {
		Found            bool                     `json:"found"`
		Inventory        protocol.HangarInventory `json:"inventory"`
		InventoryVersion int                      `json:"inventory_version"`
	}
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return protocol.HangarInventory{}, false, NewClassifiedFailure(FailureClassDecodeFailed, err)
	}
	if !decoded.Found {
		return protocol.HangarInventory{}, false, nil
	}
	if decoded.InventoryVersion > 0 {
		decoded.Inventory.InventoryVersion = decoded.InventoryVersion
	}
	return decoded.Inventory, true, nil
}

func (s *RailsStore) StoreHangarInventory(identity protocol.PlayerDataIdentity, inventory protocol.HangarInventory, expectedVersion int) (protocol.HangarInventory, error) {
	if identity.IdentityKind != IdentityKindAuthenticatedAccount {
		return protocol.HangarInventory{}, NewClassifiedFailure(FailureClassInvalidResponse, errors.New("identity_kind must be authenticated_account"))
	}
	if identity.AccountID == "" {
		return protocol.HangarInventory{}, NewClassifiedFailure(FailureClassInvalidResponse, errors.New("account_id is required"))
	}
	if s.internalToken == "" {
		return protocol.HangarInventory{}, NewClassifiedFailure(FailureClassAuthentication, errors.New("internal token is required"))
	}

	body := struct {
		AccountID       string                   `json:"account_id"`
		Inventory       protocol.HangarInventory `json:"inventory"`
		ExpectedVersion *int                     `json:"expected_version,omitempty"`
	}{AccountID: identity.AccountID, Inventory: inventory}
	if expectedVersion >= 0 {
		body.ExpectedVersion = &expectedVersion
	}
	request, err := s.newJSONRequest(http.MethodPost, "/api/internal/player-data/inventory/store", body)
	if err != nil {
		return protocol.HangarInventory{}, err
	}
	response, err := s.client().Do(request)
	if err != nil {
		return protocol.HangarInventory{}, NewClassifiedFailure(FailureClassUpstreamUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusConflict {
		return protocol.HangarInventory{}, ErrInventoryConflict
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return protocol.HangarInventory{}, NewClassifiedFailure(FailureClassUnexpectedStatus, errors.New("unexpected status"))
	}

	var decoded struct {
		Inventory        protocol.HangarInventory `json:"inventory"`
		InventoryVersion int                      `json:"inventory_version"`
	}
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return protocol.HangarInventory{}, NewClassifiedFailure(FailureClassDecodeFailed, err)
	}
	if decoded.InventoryVersion > 0 {
		decoded.Inventory.InventoryVersion = decoded.InventoryVersion
	}
	return decoded.Inventory, nil
}
