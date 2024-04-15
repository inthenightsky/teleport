// Code generated by _gen/main.go DO NOT EDIT
/*
Copyright 2015-2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"context"
	"fmt"

	apitypes "github.com/gravitational/teleport/api/types"
	
	"github.com/gravitational/teleport/integrations/lib/backoff"
	"github.com/gravitational/trace"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jonboulle/clockwork"

	"github.com/gravitational/teleport/integrations/terraform/tfschema"
)

// resourceTeleportUserType is the resource metadata type
type resourceTeleportUserType struct{}

// resourceTeleportUser is the resource
type resourceTeleportUser struct {
	p Provider
}

// GetSchema returns the resource schema
func (r resourceTeleportUserType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfschema.GenSchemaUserV2(ctx)
}

// NewResource creates the empty resource
func (r resourceTeleportUserType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceTeleportUser{
		p: *(p.(*Provider)),
	}, nil
}

// Create creates the User
func (r resourceTeleportUser) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var err error
	if !r.p.IsConfigured(resp.Diagnostics) {
		return
	}

	var plan types.Object
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := &apitypes.UserV2{}
	diags = tfschema.CopyUserV2FromTerraform(ctx, plan, user)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	
	userResource := user

	err = userResource.CheckAndSetDefaults()
	if err != nil {
		resp.Diagnostics.Append(diagFromWrappedErr("Error setting User defaults", trace.Wrap(err), "user"))
		return
	}

	id := userResource.Metadata.Name

	_, err = r.p.Client.GetUser(ctx, id, false)
	if !trace.IsNotFound(err) {
		if err == nil {
			existErr := fmt.Sprintf("User exists in Teleport. Either remove it (tctl rm user/%v)"+
				" or import it to the existing state (terraform import teleport_user.%v %v)", id, id, id)

			resp.Diagnostics.Append(diagFromErr("User exists in Teleport", trace.Errorf(existErr)))
			return
		}

		resp.Diagnostics.Append(diagFromWrappedErr("Error reading User", trace.Wrap(err), "user"))
		return
	}

	_, err = r.p.Client.CreateUser(ctx, userResource)
	if err != nil {
		resp.Diagnostics.Append(diagFromWrappedErr("Error creating User", trace.Wrap(err), "user"))
		return
	}
		
	// Not really an inferface, just using the same name for easier templating.
	var userI apitypes.User
	tries := 0
	backoff := backoff.NewDecorr(r.p.RetryConfig.Base, r.p.RetryConfig.Cap, clockwork.NewRealClock())
	for {
		tries = tries + 1
		userI, err = r.p.Client.GetUser(ctx, id, false)
		if trace.IsNotFound(err) {
			if bErr := backoff.Do(ctx); bErr != nil {
				resp.Diagnostics.Append(diagFromWrappedErr("Error reading User", trace.Wrap(bErr), "user"))
				return
			}
			if tries >= r.p.RetryConfig.MaxTries {
				diagMessage := fmt.Sprintf("Error reading User (tried %d times) - state outdated, please import resource", tries)
				resp.Diagnostics.AddError(diagMessage, "user")
			}
			continue
		}
		break
	}

	if err != nil {
		resp.Diagnostics.Append(diagFromWrappedErr("Error reading User", trace.Wrap(err), "user"))
		return
	}

	userResource, ok := userI.(*apitypes.UserV2)
	if !ok {
		resp.Diagnostics.Append(diagFromWrappedErr("Error reading User", trace.Errorf("Can not convert %T to UserV2", userI), "user"))
		return
	}
	user = userResource

	diags = tfschema.CopyUserV2ToTerraform(ctx, user, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Attrs["id"] = types.String{Value: user.Metadata.Name}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads teleport User
func (r resourceTeleportUser) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state types.Object
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var id types.String
	diags = req.State.GetAttribute(ctx, path.Root("metadata").AtName("name"), &id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userI, err := r.p.Client.GetUser(ctx, id.Value, false)
	if trace.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.Append(diagFromWrappedErr("Error reading User", trace.Wrap(err), "user"))
		return
	}
	
	user := userI.(*apitypes.UserV2)
	diags = tfschema.CopyUserV2ToTerraform(ctx, user, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates teleport User
func (r resourceTeleportUser) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	if !r.p.IsConfigured(resp.Diagnostics) {
		return
	}

	var plan types.Object
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := &apitypes.UserV2{}
	diags = tfschema.CopyUserV2FromTerraform(ctx, plan, user)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	userResource := user


	if err := userResource.CheckAndSetDefaults(); err != nil {
		resp.Diagnostics.Append(diagFromWrappedErr("Error updating User", err, "user"))
		return
	}
	name := userResource.Metadata.Name

	userBefore, err := r.p.Client.GetUser(ctx, name, false)
	if err != nil {
		resp.Diagnostics.Append(diagFromWrappedErr("Error reading User", err, "user"))
		return
	}

	_, err = r.p.Client.UpsertUser(ctx, userResource)
	if err != nil {
		resp.Diagnostics.Append(diagFromWrappedErr("Error updating User", err, "user"))
		return
	}
		
	// Not really an inferface, just using the same name for easier templating.
	var userI apitypes.User

	tries := 0
	backoff := backoff.NewDecorr(r.p.RetryConfig.Base, r.p.RetryConfig.Cap, clockwork.NewRealClock())
	for {
		tries = tries + 1
		userI, err = r.p.Client.GetUser(ctx, name, false)
		if err != nil {
			resp.Diagnostics.Append(diagFromWrappedErr("Error reading User", err, "user"))
			return
		}
		if userBefore.GetMetadata().Revision != userI.GetMetadata().Revision || false {
			break
		}

		if err := backoff.Do(ctx); err != nil {
			resp.Diagnostics.Append(diagFromWrappedErr("Error reading User", trace.Wrap(err), "user"))
			return
		}
		if tries >= r.p.RetryConfig.MaxTries {
			diagMessage := fmt.Sprintf("Error reading User (tried %d times) - state outdated, please import resource", tries)
			resp.Diagnostics.AddError(diagMessage, "user")
			return
		}
	}

	userResource, ok := userI.(*apitypes.UserV2)
	if !ok {
		resp.Diagnostics.Append(diagFromWrappedErr("Error reading User", trace.Errorf("Can not convert %T to UserV2", userI), "user"))
		return
	}
	diags = tfschema.CopyUserV2ToTerraform(ctx, user, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes Teleport User
func (r resourceTeleportUser) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var id types.String
	diags := req.State.GetAttribute(ctx, path.Root("metadata").AtName("name"), &id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.p.Client.DeleteUser(ctx, id.Value)
	if err != nil {
		resp.Diagnostics.Append(diagFromWrappedErr("Error deleting UserV2", trace.Wrap(err), "user"))
		return
	}

	resp.State.RemoveResource(ctx)
}

// ImportState imports User state
func (r resourceTeleportUser) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	user, err := r.p.Client.GetUser(ctx, req.ID, false)
	if err != nil {
		resp.Diagnostics.Append(diagFromWrappedErr("Error reading User", trace.Wrap(err), "user"))
		return
	}

	
	userResource := user.(*apitypes.UserV2)

	var state types.Object

	diags := resp.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = tfschema.CopyUserV2ToTerraform(ctx, userResource, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := userResource.GetName()

	state.Attrs["id"] = types.String{Value: id}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}