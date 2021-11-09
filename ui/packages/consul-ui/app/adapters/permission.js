import Adapter from './application';
import { inject as service } from '@ember/service';

export default class PermissionAdapter extends Adapter {
  @service('env') env;
  @service('settings') settings;

  requestForAuthorize(request, { dc, ns, resources = [], index }) {
    // the authorize endpoint is slightly different to all others in that it
    // ignores an ns parameter, but accepts a Namespace property on each
    // resource. Here we hide this difference from the rest of the app as
    // currently we never need to ask for permissions/resources for multiple
    // different namespaces in one call so here we use the ns param and add
    // this to the resources instead of passing through on the queryParameter
    if (this.env.var('CONSUL_NSPACES_ENABLED')) {
      resources = resources.map(item => ({ ...item, Namespace: ns }));
    }
    return request`
      POST /v1/internal/acl/authorize?${{ dc, index }}

      ${resources}
    `;
  }

  authorize(store, type, id, snapshot) {
    return this.rpc(
      async (adapter, request, serialized, unserialized) => {
        // the authorize endpoint does not automatically take into account the
        // default namespace of the token on the backend. This means that we
        // need to add the default namespace of the token on the frontend
        // instead. Decided this is the best place for it as its almost hidden
        // from the rest of the app so from an app eng point of view it almost
        // feels like it does happen on the backend.
        const nspacesEnabled = this.env.var('CONSUL_NSPACES_ENABLED');
        if(nspacesEnabled) {
          const token = await this.settings.findBySlug('token');
          if(typeof serialized.ns === 'undefined' || serialized.ns.length === 0) {
            serialized.ns = token.Namespace;
          }
        }
        return adapter.requestForAuthorize(request, serialized);
      },
      function(serializer, respond, serialized, unserialized) {
        // Completely skip the serializer here
        return respond(function(headers, body) {
          return body;
        });
      },
      snapshot,
      type.modelName
    );
  }
}