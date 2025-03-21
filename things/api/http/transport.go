// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MainfluxLabs/mainflux"
	log "github.com/MainfluxLabs/mainflux/logger"
	"github.com/MainfluxLabs/mainflux/pkg/apiutil"
	"github.com/MainfluxLabs/mainflux/pkg/errors"
	"github.com/MainfluxLabs/mainflux/pkg/uuid"
	"github.com/MainfluxLabs/mainflux/things"
	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(tracer opentracing.Tracer, svc things.Service, logger log.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, encodeError)),
	}

	r := bone.New()

	r.Post("/groups/:id/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_things")(createThingsEndpoint(svc)),
		decodeCreateThings,
		encodeResponse,
		opts...,
	))

	r.Patch("/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_things")(removeThingsEndpoint(svc)),
		decodeRemoveThings,
		encodeResponse,
		opts...,
	))

	r.Patch("/things/:id/key", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_key")(updateKeyEndpoint(svc)),
		decodeUpdateKey,
		encodeResponse,
		opts...,
	))

	r.Put("/things/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_thing")(updateThingEndpoint(svc)),
		decodeUpdateThing,
		encodeResponse,
		opts...,
	))

	r.Put("/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_things_metadata")(updateThingsMetadataEndpoint(svc)),
		decodeUpdateThings,
		encodeResponse,
		opts...,
	))

	r.Delete("/things/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_thing")(removeThingEndpoint(svc)),
		decodeRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/metadata", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_metadata_by_key")(viewMetadataByKeyEndpoint(svc)),
		decodeViewMetadata,
		encodeResponse,
		opts...,
	))

	r.Get("/things/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_thing")(viewThingEndpoint(svc)),
		decodeRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/things/:id/profiles", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_profile_by_thing")(viewProfileByThingEndpoint(svc)),
		decodeRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_things")(listThingsEndpoint(svc)),
		decodeList,
		encodeResponse,
		opts...,
	))

	r.Post("/things/search", kithttp.NewServer(
		kitot.TraceServer(tracer, "search_things")(listThingsEndpoint(svc)),
		decodeListByMetadata,
		encodeResponse,
		opts...,
	))

	r.Post("/groups/:id/profiles", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_profiles")(createProfilesEndpoint(svc)),
		decodeCreateProfiles,
		encodeResponse,
		opts...,
	))

	r.Patch("/profiles", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_profiles")(removeProfilesEndpoint(svc)),
		decodeRemoveProfiles,
		encodeResponse,
		opts...,
	))

	r.Put("/profiles/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_profile")(updateProfileEndpoint(svc)),
		decodeUpdateProfile,
		encodeResponse,
		opts...,
	))

	r.Delete("/profiles/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_profile")(removeProfileEndpoint(svc)),
		decodeRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/profiles/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_profile")(viewProfileEndpoint(svc)),
		decodeRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/profiles/:id/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_things_by_profile")(listThingsByProfileEndpoint(svc)),
		decodeListByID,
		encodeResponse,
		opts...,
	))

	r.Get("/profiles", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_profiles")(listProfilesEndpoint(svc)),
		decodeList,
		encodeResponse,
		opts...,
	))

	r.Post("/orgs/:id/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_groups")(createGroupsEndpoint(svc)),
		decodeCreateGroups,
		encodeResponse,
		opts...,
	))

	r.Get("/groups/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_group")(viewGroupEndpoint(svc)),
		decodeRequest,
		encodeResponse,
		opts...,
	))

	r.Put("/groups/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_group")(updateGroupEndpoint(svc)),
		decodeUpdateGroup,
		encodeResponse,
		opts...,
	))

	r.Delete("/groups/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_group")(removeGroupEndpoint(svc)),
		decodeRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_groups")(listGroupsEndpoint(svc)),
		decodeList,
		encodeResponse,
		opts...,
	))

	r.Get("/orgs/:id/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_groups_by_org")(listGroupsByOrgEndpoint(svc)),
		decodeListByID,
		encodeResponse,
		opts...,
	))

	r.Patch("/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_groups")(removeGroupsEndpoint(svc)),
		decodeRemoveGroups,
		encodeResponse,
		opts...,
	))

	r.Get("/orgs/:id/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_things_by_org")(listThingsByOrgEndpoint(svc)),
		decodeListByID,
		encodeResponse,
		opts...,
	))

	r.Get("/groups/:id/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_things_by_group")(listThingsByGroupEndpoint(svc)),
		decodeListByID,
		encodeResponse,
		opts...,
	))

	r.Get("/things/:id/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_group_by_thing")(viewGroupByThingEndpoint(svc)),
		decodeRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/orgs/:id/profiles", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_profiles_by_org")(listProfilesByOrgEndpoint(svc)),
		decodeListByID,
		encodeResponse,
		opts...,
	))

	r.Get("/groups/:id/profiles", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_profiles_by_group")(listProfilesByGroupEndpoint(svc)),
		decodeListByID,
		encodeResponse,
		opts...,
	))

	r.Get("/profiles/:id/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_group_by_profile")(viewGroupByProfileEndpoint(svc)),
		decodeRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/groups/:id/members", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_roles_by_group")(createRolesByGroupEndpoint(svc)),
		decodeGroupRoles,
		encodeResponse,
		opts...,
	))

	r.Get("/groups/:id/members", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_roles_by_group")(listRolesByGroupEndpoint(svc)),
		decodeListByID,
		encodeResponse,
		opts...,
	))

	r.Put("/groups/:id/members", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_roles_by_group")(updateRolesByGroupEndpoint(svc)),
		decodeGroupRoles,
		encodeResponse,
		opts...,
	))

	r.Patch("/groups/:id/members", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_roles_by_group")(removeRolesByGroupEndpoint(svc)),
		decodeRemoveGroupRoles,
		encodeResponse,
		opts...,
	))

	r.Get("/backup", kithttp.NewServer(
		kitot.TraceServer(tracer, "backup")(backupEndpoint(svc)),
		decodeBackup,
		encodeResponse,
		opts...,
	))

	r.Post("/restore", kithttp.NewServer(
		kitot.TraceServer(tracer, "restore")(restoreEndpoint(svc)),
		decodeRestore,
		encodeResponse,
		opts...,
	))

	r.Post("/identify", kithttp.NewServer(
		kitot.TraceServer(tracer, "identify")(identifyEndpoint(svc)),
		decodeIdentify,
		encodeResponse,
		opts...,
	))

	r.GetFunc("/health", mainflux.Health("things"))
	r.Handle("/metrics", promhttp.Handler())

	return r
}

func decodeCreateThings(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := createThingsReq{
		token:   apiutil.ExtractBearerToken(r),
		groupID: bone.GetValue(r, apiutil.IDKey),
	}
	if err := json.NewDecoder(r.Body).Decode(&req.Things); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateThing(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := updateThingReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, apiutil.IDKey),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateKey(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := updateKeyReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, apiutil.IDKey),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeViewMetadata(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewMetadataReq{
		key: apiutil.ExtractThingKey(r),
	}

	return req, nil
}

func decodeCreateProfiles(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := createProfilesReq{
		token:   apiutil.ExtractBearerToken(r),
		groupID: bone.GetValue(r, apiutil.IDKey),
	}
	if err := json.NewDecoder(r.Body).Decode(&req.Profiles); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateProfile(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := updateProfileReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, apiutil.IDKey),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeRemoveProfiles(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := removeProfilesReq{
		token: apiutil.ExtractBearerToken(r),
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := resourceReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, apiutil.IDKey),
	}

	return req, nil
}

func decodeList(_ context.Context, r *http.Request) (interface{}, error) {
	pm, err := apiutil.BuildPageMetadata(r)
	if err != nil {
		return nil, err
	}

	req := listResourcesReq{
		token:        apiutil.ExtractBearerToken(r),
		pageMetadata: pm,
	}

	return req, nil
}

func decodeListByMetadata(_ context.Context, r *http.Request) (interface{}, error) {
	req := listResourcesReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req.pageMetadata); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeListByID(_ context.Context, r *http.Request) (interface{}, error) {
	pm, err := apiutil.BuildPageMetadata(r)
	if err != nil {
		return nil, err
	}

	req := listByIDReq{
		token:        apiutil.ExtractBearerToken(r),
		id:           bone.GetValue(r, apiutil.IDKey),
		pageMetadata: pm,
	}

	return req, nil
}

func decodeCreateGroups(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := createGroupsReq{
		token: apiutil.ExtractBearerToken(r),
		orgID: bone.GetValue(r, apiutil.IDKey),
	}
	if err := json.NewDecoder(r.Body).Decode(&req.Groups); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateGroup(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := updateGroupReq{
		id:    bone.GetValue(r, apiutil.IDKey),
		token: apiutil.ExtractBearerToken(r),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeRemoveGroups(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := removeGroupsReq{
		token: apiutil.ExtractBearerToken(r),
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeRemoveThings(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := removeThingsReq{
		token: apiutil.ExtractBearerToken(r),
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateThings(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := updateThingsReq{
		token: apiutil.ExtractBearerToken(r),
	}

	if err := json.NewDecoder(r.Body).Decode(&req.Things); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeBackup(_ context.Context, r *http.Request) (interface{}, error) {
	req := backupReq{token: apiutil.ExtractBearerToken(r)}

	return req, nil
}

func decodeRestore(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := restoreReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeIdentify(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := identifyReq{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeGroupRoles(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := groupRolesReq{
		token:   apiutil.ExtractBearerToken(r),
		groupID: bone.GetValue(r, apiutil.IDKey),
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeRemoveGroupRoles(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), apiutil.ContentTypeJSON) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := removeGroupRolesReq{
		token:   apiutil.ExtractBearerToken(r),
		groupID: bone.GetValue(r, apiutil.IDKey),
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, err)
	}

	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", apiutil.ContentTypeJSON)

	if ar, ok := response.(apiutil.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}

		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch {
	case err == apiutil.ErrBearerToken,
		err == apiutil.ErrBearerKey:
		w.WriteHeader(http.StatusUnauthorized)
	case errors.Contains(err, apiutil.ErrUnsupportedContentType):
		w.WriteHeader(http.StatusUnsupportedMediaType)
	case errors.Contains(err, apiutil.ErrInvalidQueryParams),
		errors.Contains(err, apiutil.ErrMalformedEntity),
		err == apiutil.ErrNameSize,
		err == apiutil.ErrEmptyList,
		err == apiutil.ErrMissingID,
		err == apiutil.ErrMissingThingID,
		err == apiutil.ErrMissingProfileID,
		err == apiutil.ErrMissingGroupID,
		err == apiutil.ErrMissingMemberID,
		err == apiutil.ErrMissingOrgID,
		err == apiutil.ErrLimitSize,
		err == apiutil.ErrOffsetSize,
		err == apiutil.ErrInvalidOrder,
		err == apiutil.ErrInvalidDirection,
		err == apiutil.ErrInvalidIDFormat,
		err == apiutil.ErrInvalidRole:
		w.WriteHeader(http.StatusBadRequest)
	case errors.Contains(err, errors.ErrScanMetadata):
		w.WriteHeader(http.StatusUnprocessableEntity)
	case errors.Contains(err, uuid.ErrGeneratingID):
		w.WriteHeader(http.StatusInternalServerError)
	default:
		apiutil.EncodeError(err, w)
	}

	apiutil.WriteErrorResponse(err, w)
}
