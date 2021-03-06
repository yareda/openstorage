package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/libopenstorage/openstorage/api"
	client "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/cluster"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	"github.com/libopenstorage/openstorage/pkg/auth"
)

const (
	nodeOkMsg    = "Node status OK"
	nodeNotOkMsg = "Node status not OK"
)

type clusterApi struct {
	restBase
}

func newClusterAPI() restServer {
	return &clusterApi{
		restBase: restBase{
			version: cluster.APIVersion,
			name:    "Cluster API",
		},
	}
}

// SetupRoutesWithAuth sets security routes to the server router as well as
// non security routes that must not be secured in any exists.
func (c *clusterApi) SetupRoutesWithAuth(router *mux.Router, authenticators map[string]auth.Authenticator) (*mux.Router, error) {
	secureRoutes := c.SecureRoutes()
	nonSecureRoutes := c.Routes()

	routeMap := make(map[string]*Route)

	// fill map with non-secure routes
	for _, route := range nonSecureRoutes {
		routeMap[route.GetPath()+route.GetVerb()] = route
	}

	// Remove routes that shall not be secured
	for _, route := range secureRoutes {
		delete(routeMap, route.GetPath()+route.GetVerb())
	}

	securityMiddleware := newSecurityMiddleware(authenticators)

	for _, route := range secureRoutes {
		router.Methods(route.GetVerb()).Path(route.GetPath()).HandlerFunc(securityMiddleware(route.fn))
	}

	// Put all non-secured routes on router
	for _, route := range routeMap {
		router.Methods(route.GetVerb()).Path(route.GetPath()).HandlerFunc(route.GetFn())
	}

	return router, nil
}

func (c *clusterApi) String() string {
	return c.name
}

// swagger:operation GET /cluster/enumerate cluster enumerateCluster
//
// Lists cluster Nodes.
//
// This will return the entire cluster object and it's nodes.
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: current cluster state
//      schema:
//         type: array
//         items:
//            $ref: '#/definitions/Cluster'
func (c *clusterApi) enumerate(w http.ResponseWriter, r *http.Request) {
	method := "enumerate"
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	cluster, err := inst.Enumerate()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cluster)
}

func (c *clusterApi) setSize(w http.ResponseWriter, r *http.Request) {
	method := "set size"
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := r.URL.Query()

	size := params["size"]
	if size == nil {
		c.sendError(c.name, method, w, "Missing size param", http.StatusBadRequest)
		return
	}

	sz, _ := strconv.Atoi(size[0])

	err = inst.SetSize(sz)

	clusterResponse := &api.ClusterResponse{Error: err.Error()}
	json.NewEncoder(w).Encode(clusterResponse)
}

// swagger:operation GET /cluster/inspect/{id} cluster inspectNode
//
// Inspect cluster Nodes.
//
// This will return the requested node object
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get node with
//   required: true
//   type: integer
// responses:
//   '200':
//      description: a node
//      schema:
//       $ref: '#/definitions/Node'
func (c *clusterApi) inspect(w http.ResponseWriter, r *http.Request) {
	method := "inspect"
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	nodeID, ok := vars["id"]

	if !ok || nodeID == "" {
		c.sendError(c.name, method, w, "Missing id param", http.StatusBadRequest)
		return
	}

	if nodeStats, err := inst.Inspect(nodeID); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
	} else {
		json.NewEncoder(w).Encode(nodeStats)
	}
}

func (c *clusterApi) enableGossip(w http.ResponseWriter, r *http.Request) {
	method := "enablegossip"

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	inst.EnableUpdates()

	clusterResponse := &api.ClusterResponse{}
	json.NewEncoder(w).Encode(clusterResponse)
}

func (c *clusterApi) disableGossip(w http.ResponseWriter, r *http.Request) {
	method := "disablegossip"

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	inst.DisableUpdates()

	clusterResponse := &api.ClusterResponse{}
	json.NewEncoder(w).Encode(clusterResponse)
}

func (c *clusterApi) gossipState(w http.ResponseWriter, r *http.Request) {
	method := "gossipState"

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := inst.GetGossipState()
	json.NewEncoder(w).Encode(resp)
}

// swagger:operation GET /cluster/getnodeidfromip/{idip} cluster GetNodeIdFromIp
//
// this will return the node ID for the given node IP
//
// ---
// produces:
// - application/json
// parameters:
// - name: idip
//   in: path
//   description: cluster node ip or id
//   required: true
//   type: string
// responses:
//   '200':
//      description: cluster node ID
//      schema:
//         type: string
func (c *clusterApi) getNodeIdFromIp(w http.ResponseWriter, r *http.Request) {
	method := "getnodeidfromip"
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	nodeIP, ok := vars["idip"]

	if !ok || nodeIP == "" {
		c.sendError(c.name, method, w, "Missing id param", http.StatusBadRequest)
		return
	}

	if nodeID, err := inst.GetNodeIdFromIp(nodeIP); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
	} else {
		json.NewEncoder(w).Encode(nodeID)
	}
}

// swagger:operation GET /cluster/status cluster status
//
// this will return the cluster status.
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: cluster status
//      schema:
//         type: string
func (c *clusterApi) status(w http.ResponseWriter, r *http.Request) {
	method := "status"

	inst, err := clustermanager.Inst()

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	cluster, err := inst.Enumerate()

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cluster.Status)
}

func nodeStatusIntl() (api.Status, error) {
	inst, err := clustermanager.Inst()
	if err != nil {
		return api.Status_STATUS_NONE, err
	}

	resp, err := inst.NodeStatus()
	if err != nil {
		return api.Status_STATUS_NONE, err
	}

	return resp, nil
}

// swagger:operation GET /cluster/nodestatus node nodeStatus
//
// This will return the node status .
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: node status of responding node.
//      schema:
//         type: string
func (c *clusterApi) nodeStatus(w http.ResponseWriter, r *http.Request) {
	method := "nodeStatus"

	st, err := nodeStatusIntl()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(st)
}

// swagger:operation GET /cluster/nodehealth node nodeHealth
//
// This will return node health.
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: node health of responding node.
//      schema:
//         type: string
func (c *clusterApi) nodeHealth(w http.ResponseWriter, r *http.Request) {
	method := "nodeHealth"

	st, err := nodeStatusIntl()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	if st != api.Status_STATUS_OK {
		err = fmt.Errorf("%s (%s)", nodeNotOkMsg, api.Status_name[int32(st)])
		c.sendError(c.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Write([]byte(nodeOkMsg + "\n"))
}

// swagger:operation GET /cluster/peerstatus node peerStatus
//
// This will return the peer node status
//
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: query
//   description: id of the node we want to check.
//   required: true
//   type: integer
// responses:
//   '200':
//      description: node status of requested node
//      schema:
//         type: string
func (c *clusterApi) peerStatus(w http.ResponseWriter, r *http.Request) {
	method := "peerStatus"

	params := r.URL.Query()
	listenerName := params["name"]
	if len(listenerName) == 0 || listenerName[0] == "" {
		c.sendError(c.name, method, w, "Missing id param", http.StatusBadRequest)
		return
	}
	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := inst.PeerStatus(listenerName[0])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)

}

// swagger:operation DELETE /cluster/{id} cluster deleteNode
//
// This will delete a node from the cluster
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get node with
//   required: true
//   type: integer
// - name: forceRemove
//   in: query
//   description: forceRemove node
//   required: false
//   type: boolean
// responses:
//   '200':
//      description: delete node success
//      schema:
//         type: string
func (c *clusterApi) delete(w http.ResponseWriter, r *http.Request) {
	method := "delete"

	params := r.URL.Query()

	nodeID := params["id"]
	if nodeID == nil {
		c.sendError(c.name, method, w, "Missing id param", http.StatusBadRequest)
		return
	}

	forceRemoveParam := params["forceRemove"]
	forceRemove := false
	if forceRemoveParam != nil {
		var err error
		forceRemove, err = strconv.ParseBool(forceRemoveParam[0])
		if err != nil {
			c.sendError(c.name, method, w, "Invalid forceRemove Option: "+
				forceRemoveParam[0], http.StatusBadRequest)
			return
		}
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	nodes := make([]api.Node, 0)
	for _, id := range nodeID {
		nodes = append(nodes, api.Node{Id: id})
	}

	clusterResponse := &api.ClusterResponse{}

	err = inst.Remove(nodes, forceRemove)
	if err != nil {
		clusterResponse.Error = err.Error()
	}
	json.NewEncoder(w).Encode(clusterResponse)
}

// swagger:operation PUT /cluster/{id} cluster shutdownNode
//
// This will shutdown a node (Not Implemented)
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get node with
//   required: true
//   type: integer
// responses:
//   '200':
//      description: shutdown success
//      schema:
//         type: string
func (c *clusterApi) shutdown(w http.ResponseWriter, r *http.Request) {
	method := "shutdown"
	c.sendNotImplemented(w, method)
}

// swagger:operation GET /cluster/versions cluster enumerateVersions
//
// Lists API Versions supported by this cluster
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: Supported versions
//      schema:
//         type: array
//         items:
//            type: string
func (c *clusterApi) versions(w http.ResponseWriter, r *http.Request) {
	versions := []string{
		cluster.APIVersion,
		// Update supported versions by adding them here
	}
	json.NewEncoder(w).Encode(versions)
}

// swagger:operation GET /cluster/alerts/{resource} cluster enumerateAlerts
//
// This will return a list of alerts for the requested resource
//
// ---
// produces:
// - application/json
// parameters:
// - name: resource
//   in: path
//   description: |
//    Resourcetype to get alerts with.
//    0: All
//    1: Volume
//    2: Node
//    3: Cluster
//    4: Drive
//   required: true
//   type: integer
// responses:
//   '200':
//      description: Alerts object
//      schema:
//       $ref: '#/definitions/Alerts'
func (c *clusterApi) enumerateAlerts(w http.ResponseWriter, r *http.Request) {
	method := "enumerateAlerts"

	params := r.URL.Query()

	var (
		resourceType api.ResourceType
		err          error
		tS, tE       time.Time
	)
	vars := mux.Vars(r)
	resource, ok := vars["resource"]
	if ok {
		resourceType, err = handleResourceType(resource)
		if err != nil {
			c.sendError(c.name, method, w, "Invalid resource param", http.StatusBadRequest)
			return
		}
	} else {
		resourceType = api.ResourceType_RESOURCE_TYPE_NONE
	}

	timeStart := params["timestart"]
	if timeStart != nil {
		tS, err = time.Parse(api.TimeLayout, timeStart[0])
		if err != nil {
			c.sendError(c.name, method, w, "Invalid timestart param", http.StatusBadRequest)
			return
		}
	}

	timeEnd := params["timeend"]
	if timeEnd != nil {
		tS, err = time.Parse(api.TimeLayout, timeEnd[0])
		if err != nil {
			c.sendError(c.name, method, w, "Invalid timeend param", http.StatusBadRequest)
			return
		}
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	alerts, err := inst.EnumerateAlerts(tS, tE, resourceType)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(alerts)
}

// swagger:operation DELETE /cluster/alerts/{resource}/{id} cluster deleteAlert
//
// This delete clear alert {id} with resourcetype {resource}
//
// ---
// produces:
// - application/json
// parameters:
// - name: resource
//   in: path
//   description: |
//    resourcetype to get alerts with.
//    0: All
//    1: Volume
//    2: Node
//    3: Cluster
//    4: Drive
//   required: true
//   type: integer
// - name: id
//   in: path
//   description: id to get alerts with
//   required: true
//   type: integer
// responses:
//   '200':
//      description: Alerts object
//      schema:
//       type: string
func (c *clusterApi) eraseAlert(w http.ResponseWriter, r *http.Request) {
	method := "eraseAlert"

	resourceType, alertId, err := c.getAlertParams(w, r, method)
	if err != nil {
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.EraseAlert(resourceType, alertId)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode("Successfully erased Alert")
}

func (c *clusterApi) getAlertParams(w http.ResponseWriter, r *http.Request, method string) (api.ResourceType, int64, error) {
	var (
		resourceType api.ResourceType
		alertId      int64
		err          error
	)
	returnErr := fmt.Errorf("Invalid param")

	vars := mux.Vars(r)
	resource, ok := vars["resource"]
	if ok {
		resourceType, err = handleResourceType(resource)
	}

	if err != nil || !ok {
		c.sendError(c.name, method, w, "Missing/Invalid resource param", http.StatusBadRequest)
		return api.ResourceType_RESOURCE_TYPE_NONE, 0, returnErr

	}

	vars = mux.Vars(r)
	id, ok := vars["id"]
	if ok {
		alertId, err = strconv.ParseInt(id, 10, 64)
	}

	if err != nil || !ok {
		c.sendError(c.name, method, w, "Missing/Invalid id param", http.StatusBadRequest)
		return api.ResourceType_RESOURCE_TYPE_NONE, 0, returnErr
	}
	return resourceType, alertId, nil
}

func (c *clusterApi) sendNotImplemented(w http.ResponseWriter, method string) {
	c.sendError(c.name, method, w, "Not implemented.", http.StatusNotImplemented)
}

func clusterVersion(route, version string) string {
	return "/" + version + "/" + route
}

func clusterSecretPath(route, version string) string {
	return clusterPath("/secrets"+route, version)
}

func clusterPath(route, version string) string {
	return clusterVersion("cluster"+route, version)
}

func clusterPairPath(route, version string) string {
	return clusterPath(client.PairPath+route, version)
}

func handleResourceType(resource string) (api.ResourceType, error) {
	resource = strings.ToLower(resource)
	switch resource {
	case "volume":
		return api.ResourceType_RESOURCE_TYPE_VOLUME, nil
	case "node":
		return api.ResourceType_RESOURCE_TYPE_NODE, nil
	case "cluster":
		return api.ResourceType_RESOURCE_TYPE_CLUSTER, nil
	case "drive":
		return api.ResourceType_RESOURCE_TYPE_DRIVE, nil
	default:
		resourceType, err := strconv.ParseInt(resource, 10, 64)
		if err == nil {
			if _, ok := api.ResourceType_name[int32(resourceType)]; ok {
				return api.ResourceType(resourceType), nil
			}
		}
		return api.ResourceType_RESOURCE_TYPE_NONE, fmt.Errorf("Invalid resource type")
	}
}

func (c *clusterApi) createPair(w http.ResponseWriter, r *http.Request) {
	pairRequest := &api.ClusterPairCreateRequest{}
	method := "createPair"

	if err := json.NewDecoder(r.Body).Decode(pairRequest); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := inst.CreatePair(pairRequest)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *clusterApi) processPair(w http.ResponseWriter, r *http.Request) {
	processPairRequest := &api.ClusterPairProcessRequest{}
	method := "processPair"

	if err := json.NewDecoder(r.Body).Decode(processPairRequest); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := inst.ProcessPairRequest(processPairRequest)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *clusterApi) enumeratePairs(w http.ResponseWriter, r *http.Request) {
	method := "enumeratePairs"

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := inst.EnumeratePairs()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *clusterApi) getPair(w http.ResponseWriter, r *http.Request) {
	method := "getPair"

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		c.sendError(c.name, method, w, "id required for GET Pair request", http.StatusBadRequest)
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := inst.GetPair(id)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (c *clusterApi) refreshPair(w http.ResponseWriter, r *http.Request) {
	method := "refreshPair"

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		c.sendError(c.name, method, w, "id required for refresh Pair request", http.StatusBadRequest)
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.RefreshPair(id)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode("Successfully refreshed cluster pair")
}

func (c *clusterApi) validatePair(w http.ResponseWriter, r *http.Request) {
	method := "validatePair"

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		c.sendError(c.name, method, w, "id required for validate pair request", http.StatusBadRequest)
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.ValidatePair(id)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode("Successfully validated cluster pair")
}

func (c *clusterApi) deletePair(w http.ResponseWriter, r *http.Request) {
	method := "deletePair"

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		c.sendError(c.name, method, w, "id required for DELETE Pair request", http.StatusBadRequest)
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.DeletePair(id)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode("Successfully deleted pairing with cluster")
}

func (c *clusterApi) getPairToken(w http.ResponseWriter, r *http.Request) {
	method := "getPairToken"

	var err error
	reset := false
	params := r.URL.Query()
	resetString := params["reset"]
	if resetString != nil {
		reset, err = strconv.ParseBool(resetString[0])
		if err != nil {
			c.sendError(c.name, method, w, "Invalid reset parameter", http.StatusBadRequest)
			return
		}
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := inst.GetPairToken(reset)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

// Routes method returns list of all routes served by cluster API. Those can be both
// routes supposed to be secured in case if security enabled and those that don't.
func (c *clusterApi) Routes() []*Route {
	return []*Route{
		{verb: http.MethodGet, path: clusterPath("/enumerate", cluster.APIVersion), fn: c.enumerate},
		{verb: http.MethodGet, path: clusterPath("/gossipstate", cluster.APIVersion), fn: c.gossipState},
		{verb: http.MethodGet, path: clusterPath("/nodestatus", cluster.APIVersion), fn: c.nodeStatus},
		{verb: http.MethodGet, path: clusterPath("/nodehealth", cluster.APIVersion), fn: c.nodeHealth},
		{verb: http.MethodGet, path: clusterPath("/status", cluster.APIVersion), fn: c.status},
		{verb: http.MethodGet, path: clusterPath("/peerstatus", cluster.APIVersion), fn: c.peerStatus},
		{verb: http.MethodGet, path: clusterPath("/inspect/{id}", cluster.APIVersion), fn: c.inspect},

		{verb: http.MethodDelete, path: clusterPath("", cluster.APIVersion), fn: c.delete},
		{verb: http.MethodDelete, path: clusterPath("/{id}", cluster.APIVersion), fn: c.delete},

		{verb: http.MethodPut, path: clusterPath("/enablegossip", cluster.APIVersion), fn: c.enableGossip},
		{verb: http.MethodPut, path: clusterPath("/disablegossip", cluster.APIVersion), fn: c.disableGossip},
		{verb: http.MethodPut, path: clusterPath("/shutdown", cluster.APIVersion), fn: c.shutdown},
		{verb: http.MethodPut, path: clusterPath("/shutdown/{id}", cluster.APIVersion), fn: c.shutdown},

		{verb: http.MethodGet, path: clusterPath("/alerts/{resource}", cluster.APIVersion), fn: c.enumerateAlerts},
		{verb: http.MethodGet, path: "/cluster/versions", fn: c.versions},

		{verb: http.MethodDelete, path: clusterPath("/alerts/{resource}/{id}", cluster.APIVersion), fn: c.eraseAlert},
		{verb: http.MethodPut, path: clusterPairPath("", cluster.APIVersion), fn: c.createPair},
		{verb: http.MethodPost, path: clusterPairPath("", cluster.APIVersion), fn: c.processPair},
		{verb: http.MethodGet, path: clusterPairPath("", cluster.APIVersion), fn: c.enumeratePairs},
		{verb: http.MethodGet, path: clusterPairPath("/{id}", cluster.APIVersion), fn: c.getPair},
		{verb: http.MethodPut, path: clusterPairPath("/{id}", cluster.APIVersion), fn: c.refreshPair},
		{verb: http.MethodDelete, path: clusterPairPath("/{id}", cluster.APIVersion), fn: c.deletePair},
		{verb: http.MethodPut, path: clusterPairPath(client.PairValidatePath+"/{id}", cluster.APIVersion), fn: c.validatePair},
		{verb: http.MethodGet, path: clusterPath(client.PairTokenPath, cluster.APIVersion), fn: c.getPairToken},

		{verb: http.MethodGet, path: clusterPath(client.SchedPath, cluster.APIVersion), fn: c.schedPolicyEnumerate},
		{verb: http.MethodGet, path: clusterPath(client.SchedPath+"/{name}", cluster.APIVersion), fn: c.schedPolicyGet},
		{verb: http.MethodPost, path: clusterPath(client.SchedPath, cluster.APIVersion), fn: c.schedPolicyCreate},
		{verb: http.MethodPut, path: clusterPath(client.SchedPath, cluster.APIVersion), fn: c.schedPolicyUpdate},
		{verb: http.MethodDelete, path: clusterPath(client.SchedPath+"/{name}", cluster.APIVersion), fn: c.schedPolicyDelete},
		{verb: http.MethodGet, path: clusterPath(client.ObjectStorePath, cluster.APIVersion), fn: c.objectStoreInspect},
		{verb: http.MethodPost, path: clusterPath(client.ObjectStorePath, cluster.APIVersion), fn: c.objectStoreCreate},
		{verb: http.MethodPut, path: clusterPath(client.ObjectStorePath, cluster.APIVersion), fn: c.objectStoreUpdate},
		{verb: http.MethodDelete, path: clusterPath(client.ObjectStorePath+"/delete", cluster.APIVersion), fn: c.objectStoreDelete},

		{verb: http.MethodGet, path: clusterSecretPath("/verify", cluster.APIVersion), fn: c.secretLoginCheck},
		{verb: http.MethodGet, path: clusterSecretPath("", cluster.APIVersion), fn: c.getSecret},
		{verb: http.MethodPut, path: clusterSecretPath("", cluster.APIVersion), fn: c.setSecret},
		{verb: http.MethodGet, path: clusterSecretPath("/defaultsecretkey", cluster.APIVersion), fn: c.getDefaultSecretKey},
		{verb: http.MethodPut, path: clusterSecretPath("/defaultsecretkey", cluster.APIVersion), fn: c.setDefaultSecretKey},
		{verb: http.MethodPost, path: clusterSecretPath("/login", cluster.APIVersion), fn: c.secretsLogin},

		{verb: http.MethodGet, path: clusterPath(client.UriCluster, cluster.APIVersion), fn: c.getClusterConf},
		{verb: http.MethodGet, path: clusterPath(client.UriNode+"/{id}", cluster.APIVersion), fn: c.getNodeConf},
		{verb: http.MethodGet, path: clusterPath(client.UriEnumerate, cluster.APIVersion), fn: c.enumerateConf},
		{verb: http.MethodPost, path: clusterPath(client.UriCluster, cluster.APIVersion), fn: c.setClusterConf},
		{verb: http.MethodPost, path: clusterPath(client.UriNode, cluster.APIVersion), fn: c.setNodeConf},
		{verb: http.MethodDelete, path: clusterPath(client.UriNode+"/{id}", cluster.APIVersion), fn: c.delNodeConf},
		{verb: http.MethodGet, path: clusterPath("/getnodeidfromip/{idip}", cluster.APIVersion), fn: c.getNodeIdFromIp},
	}
}

// SecureRoutes return list of routes that only required to be secured.
func (c *clusterApi) SecureRoutes() []*Route {
	return []*Route{
		{verb: http.MethodGet, path: clusterPath("/enumerate", cluster.APIVersion), fn: c.enumerate},
		{verb: http.MethodGet, path: clusterPath("/gossipstate", cluster.APIVersion), fn: c.gossipState},
		{verb: http.MethodGet, path: clusterPath("/nodestatus", cluster.APIVersion), fn: c.nodeStatus},
		{verb: http.MethodGet, path: clusterPath("/status", cluster.APIVersion), fn: c.status},
		{verb: http.MethodGet, path: clusterPath("/peerstatus", cluster.APIVersion), fn: c.peerStatus},
		{verb: http.MethodGet, path: clusterPath("/inspect/{id}", cluster.APIVersion), fn: c.inspect},

		{verb: http.MethodDelete, path: clusterPath("", cluster.APIVersion), fn: c.delete},
		{verb: http.MethodDelete, path: clusterPath("/{id}", cluster.APIVersion), fn: c.delete},

		{verb: http.MethodPut, path: clusterPath("/enablegossip", cluster.APIVersion), fn: c.enableGossip},
		{verb: http.MethodPut, path: clusterPath("/disablegossip", cluster.APIVersion), fn: c.disableGossip},
		{verb: http.MethodPut, path: clusterPath("/shutdown", cluster.APIVersion), fn: c.shutdown},
		{verb: http.MethodPut, path: clusterPath("/shutdown/{id}", cluster.APIVersion), fn: c.shutdown},

		{verb: http.MethodGet, path: clusterPath("/alerts/{resource}", cluster.APIVersion), fn: c.enumerateAlerts},
		{verb: http.MethodGet, path: "/cluster/versions", fn: c.versions},

		{verb: http.MethodDelete, path: clusterPath("/alerts/{resource}/{id}", cluster.APIVersion), fn: c.eraseAlert},
		{verb: http.MethodPut, path: clusterPairPath("", cluster.APIVersion), fn: c.createPair},
		{verb: http.MethodPost, path: clusterPairPath("", cluster.APIVersion), fn: c.processPair},
		{verb: http.MethodGet, path: clusterPairPath("", cluster.APIVersion), fn: c.enumeratePairs},
		{verb: http.MethodGet, path: clusterPairPath("/{id}", cluster.APIVersion), fn: c.getPair},
		{verb: http.MethodPut, path: clusterPairPath("/{id}", cluster.APIVersion), fn: c.refreshPair},
		{verb: http.MethodDelete, path: clusterPairPath("/{id}", cluster.APIVersion), fn: c.deletePair},
		{verb: http.MethodPut, path: clusterPairPath(client.PairValidatePath+"/{id}", cluster.APIVersion), fn: c.validatePair},
		{verb: http.MethodGet, path: clusterPath(client.PairTokenPath, cluster.APIVersion), fn: c.getPairToken},

		{verb: http.MethodGet, path: clusterPath(client.SchedPath, cluster.APIVersion), fn: c.schedPolicyEnumerate},
		{verb: http.MethodGet, path: clusterPath(client.SchedPath+"/{name}", cluster.APIVersion), fn: c.schedPolicyGet},
		{verb: http.MethodPost, path: clusterPath(client.SchedPath, cluster.APIVersion), fn: c.schedPolicyCreate},
		{verb: http.MethodPut, path: clusterPath(client.SchedPath, cluster.APIVersion), fn: c.schedPolicyUpdate},
		{verb: http.MethodDelete, path: clusterPath(client.SchedPath+"/{name}", cluster.APIVersion), fn: c.schedPolicyDelete},
		{verb: http.MethodGet, path: clusterPath(client.ObjectStorePath, cluster.APIVersion), fn: c.objectStoreInspect},
		{verb: http.MethodPost, path: clusterPath(client.ObjectStorePath, cluster.APIVersion), fn: c.objectStoreCreate},
		{verb: http.MethodPut, path: clusterPath(client.ObjectStorePath, cluster.APIVersion), fn: c.objectStoreUpdate},
		{verb: http.MethodDelete, path: clusterPath(client.ObjectStorePath+"/delete", cluster.APIVersion), fn: c.objectStoreDelete},

		{verb: http.MethodGet, path: clusterSecretPath("/verify", cluster.APIVersion), fn: c.secretLoginCheck},
		{verb: http.MethodGet, path: clusterSecretPath("", cluster.APIVersion), fn: c.getSecret},
		{verb: http.MethodPut, path: clusterSecretPath("", cluster.APIVersion), fn: c.setSecret},
		{verb: http.MethodGet, path: clusterSecretPath("/defaultsecretkey", cluster.APIVersion), fn: c.getDefaultSecretKey},
		{verb: http.MethodPut, path: clusterSecretPath("/defaultsecretkey", cluster.APIVersion), fn: c.setDefaultSecretKey},
		{verb: http.MethodPost, path: clusterSecretPath("/login", cluster.APIVersion), fn: c.secretsLogin},

		{verb: http.MethodGet, path: clusterPath(client.UriCluster, cluster.APIVersion), fn: c.getClusterConf},
		{verb: http.MethodGet, path: clusterPath(client.UriNode+"/{id}", cluster.APIVersion), fn: c.getNodeConf},
		{verb: http.MethodGet, path: clusterPath(client.UriEnumerate, cluster.APIVersion), fn: c.enumerateConf},
		{verb: http.MethodPost, path: clusterPath(client.UriCluster, cluster.APIVersion), fn: c.setClusterConf},
		{verb: http.MethodPost, path: clusterPath(client.UriNode, cluster.APIVersion), fn: c.setNodeConf},
		{verb: http.MethodDelete, path: clusterPath(client.UriNode+"/{id}", cluster.APIVersion), fn: c.delNodeConf},
		{verb: http.MethodGet, path: clusterPath("/getnodeidfromip/{idip}", cluster.APIVersion), fn: c.getNodeIdFromIp},
	}
}
