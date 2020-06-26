package workloads

import schema "github.com/threefoldtech/tfexplorer/schema"

type GatewayProxy struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	Addr       string    `bson:"addr" json:"addr"`
	Port       uint32    `bson:"port" json:"port"`
	PortTLS    uint32    `bson:"port_tls" json:"port_tls"`
	PoolId     int64     `bson:"pool_id" json:"pool_id"`

	Description             string         `bson:"description" json:"description"`
	Currencies              []string       `bson:"currencies" json:"currencies"`
	SigningRequestProvision SigningRequest `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest `bson:"signing_request_delete" json:"signing_request_delete"`
	ExpirationProvisioning  schema.Date    `bson:"expiration_provisioning" json:"expiration_provisioning"`

	Json                string             `bson:"json" json:"json"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignatureFarmer     SigningSignature   `bson:"signature_farmer" json:"signature_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Metadata            string             `bson:"metadata" json:"metadata"`
	Result              Result             `bson:"result" json:"result"`
	WorkloadType        WorkloadTypeEnum   `bson:"workload_type" json:"workload_type"`
}

func (g *GatewayProxy) WorkloadID() int64 {
	return g.WorkloadId
}

func (g *GatewayProxy) GetWorkloadType() WorkloadTypeEnum {
	return g.WorkloadType
}

func (g *GatewayProxy) GetID() schema.ID {
	return g.ID
}

func (g *GatewayProxy) SetID(id schema.ID) {
	g.ID = id
}

func (g *GatewayProxy) GetJson() string {
	return g.Json
}

func (g *GatewayProxy) GetCustomerTid() int64 {
	return g.CustomerTid
}

func (g *GatewayProxy) GetCustomerSignature() string {
	return g.CustomerSignature
}

func (g *GatewayProxy) GetNextAction() NextActionEnum {
	return g.NextAction
}

func (g *GatewayProxy) SetNextAction(next NextActionEnum) {
	g.NextAction = next
}

func (g *GatewayProxy) GetSignaturesProvision() []SigningSignature {
	return g.SignaturesProvision
}

func (g *GatewayProxy) PushSignatureProvision(signature SigningSignature) {
	g.SignaturesProvision = append(g.SignaturesProvision, signature)
}

func (g *GatewayProxy) GetSignatureFarmer() SigningSignature {
	return g.SignatureFarmer
}

func (g *GatewayProxy) SetSignatureFarmer(signature SigningSignature) {
	g.SignatureFarmer = signature
}

func (g *GatewayProxy) GetSignaturesDelete() []SigningSignature {
	return g.SignaturesDelete
}

func (g *GatewayProxy) PushSignatureDelete(signature SigningSignature) {
	g.SignaturesDelete = append(g.SignaturesDelete, signature)
}

func (g *GatewayProxy) GetEpoch() schema.Date {
	return g.Epoch
}

func (g *GatewayProxy) GetMetadata() string {
	return g.Metadata
}

func (g *GatewayProxy) GetResult() Result {
	return g.Result
}

func (g *GatewayProxy) SetResult(result Result) {
	g.Result = result
}

func (g *GatewayProxy) GetDescription() string {
	return g.Description
}

func (g *GatewayProxy) GetCurrencies() []string {
	return g.Currencies
}

func (g *GatewayProxy) GetSigningRequestProvision() SigningRequest {
	return g.SigningRequestProvision
}

func (g *GatewayProxy) GetSigningRequestDelete() SigningRequest {
	return g.SigningRequestDelete
}

func (g *GatewayProxy) GetExpirationProvisioning() schema.Date {
	return g.ExpirationProvisioning
}

func (g *GatewayProxy) SetJson(json string) {
	g.Json = json
}

func (g *GatewayProxy) SetCustomerTid(tid int64) {
	g.CustomerTid = tid
}

func (g *GatewayProxy) SetCustomerSignature(signature string) {
	g.CustomerSignature = signature
}

func (g *GatewayProxy) SetEpoch(date schema.Date) {
	g.Epoch = date
}

func (g *GatewayProxy) SetMetadata(metadata string) {
	g.Metadata = metadata
}

func (g *GatewayProxy) SetDescription(description string) {
	g.Description = description
}

func (g *GatewayProxy) SetCurrencies(currencies []string) {
	g.Currencies = currencies
}

func (g *GatewayProxy) SetSigningRequestProvision(request SigningRequest) {
	g.SigningRequestProvision = request
}

func (g *GatewayProxy) SetSigningRequestDelete(request SigningRequest) {
	g.SigningRequestDelete = request
}

func (g *GatewayProxy) SetExpirationProvisioning(date schema.Date) {
	g.ExpirationProvisioning = date
}

type GatewayReverseProxy struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	Secret     string    `bson:"secret" json:"secret"`
	PoolId     int64     `bson:"pool_id" json:"pool_id"`

	Description             string         `bson:"description" json:"description"`
	Currencies              []string       `bson:"currencies" json:"currencies"`
	SigningRequestProvision SigningRequest `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest `bson:"signing_request_delete" json:"signing_request_delete"`
	ExpirationProvisioning  schema.Date    `bson:"expiration_provisioning" json:"expiration_provisioning"`

	Json                string             `bson:"json" json:"json"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignatureFarmer     SigningSignature   `bson:"signature_farmer" json:"signature_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Metadata            string             `bson:"metadata" json:"metadata"`
	Result              Result             `bson:"result" json:"result"`
	WorkloadType        WorkloadTypeEnum   `bson:"workload_type" json:"workload_type"`
}

func (g *GatewayReverseProxy) WorkloadID() int64 {
	return g.WorkloadId
}

func (g *GatewayReverseProxy) GetWorkloadType() WorkloadTypeEnum {
	return g.WorkloadType
}

func (g *GatewayReverseProxy) GetID() schema.ID {
	return g.ID
}

func (g *GatewayReverseProxy) SetID(id schema.ID) {
	g.ID = id
}

func (g *GatewayReverseProxy) GetJson() string {
	return g.Json
}

func (g *GatewayReverseProxy) GetCustomerTid() int64 {
	return g.CustomerTid
}

func (g *GatewayReverseProxy) GetCustomerSignature() string {
	return g.CustomerSignature
}

func (g *GatewayReverseProxy) GetNextAction() NextActionEnum {
	return g.NextAction
}

func (g *GatewayReverseProxy) SetNextAction(next NextActionEnum) {
	g.NextAction = next
}

func (g *GatewayReverseProxy) GetSignaturesProvision() []SigningSignature {
	return g.SignaturesProvision
}

func (g *GatewayReverseProxy) PushSignatureProvision(signature SigningSignature) {
	g.SignaturesProvision = append(g.SignaturesProvision, signature)
}

func (g *GatewayReverseProxy) GetSignatureFarmer() SigningSignature {
	return g.SignatureFarmer
}

func (g *GatewayReverseProxy) SetSignatureFarmer(signature SigningSignature) {
	g.SignatureFarmer = signature
}

func (g *GatewayReverseProxy) GetSignaturesDelete() []SigningSignature {
	return g.SignaturesDelete
}

func (g *GatewayReverseProxy) PushSignatureDelete(signature SigningSignature) {
	g.SignaturesDelete = append(g.SignaturesDelete, signature)
}

func (g *GatewayReverseProxy) GetEpoch() schema.Date {
	return g.Epoch
}

func (g *GatewayReverseProxy) GetMetadata() string {
	return g.Metadata
}

func (g *GatewayReverseProxy) GetResult() Result {
	return g.Result
}

func (g *GatewayReverseProxy) SetResult(result Result) {
	g.Result = result
}

func (g *GatewayReverseProxy) GetDescription() string {
	return g.Description
}

func (g *GatewayReverseProxy) GetCurrencies() []string {
	return g.Currencies
}

func (g *GatewayReverseProxy) GetSigningRequestProvision() SigningRequest {
	return g.SigningRequestProvision
}

func (g *GatewayReverseProxy) GetSigningRequestDelete() SigningRequest {
	return g.SigningRequestDelete
}

func (g *GatewayReverseProxy) GetExpirationProvisioning() schema.Date {
	return g.ExpirationProvisioning
}

func (g *GatewayReverseProxy) SetJson(json string) {
	g.Json = json
}

func (g *GatewayReverseProxy) SetCustomerTid(tid int64) {
	g.CustomerTid = tid
}

func (g *GatewayReverseProxy) SetCustomerSignature(signature string) {
	g.CustomerSignature = signature
}

func (g *GatewayReverseProxy) SetEpoch(date schema.Date) {
	g.Epoch = date
}

func (g *GatewayReverseProxy) SetMetadata(metadata string) {
	g.Metadata = metadata
}

func (g *GatewayReverseProxy) SetDescription(description string) {
	g.Description = description
}

func (g *GatewayReverseProxy) SetCurrencies(currencies []string) {
	g.Currencies = currencies
}

func (g *GatewayReverseProxy) SetSigningRequestProvision(request SigningRequest) {
	g.SigningRequestProvision = request
}

func (g *GatewayReverseProxy) SetSigningRequestDelete(request SigningRequest) {
	g.SigningRequestDelete = request
}

func (g *GatewayReverseProxy) SetExpirationProvisioning(date schema.Date) {
	g.ExpirationProvisioning = date
}

type GatewaySubdomain struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	IPs        []string  `bson:"ips" json:"ips"`
	PoolId     int64     `bson:"pool_id" json:"pool_id"`

	Description             string         `bson:"description" json:"description"`
	Currencies              []string       `bson:"currencies" json:"currencies"`
	SigningRequestProvision SigningRequest `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest `bson:"signing_request_delete" json:"signing_request_delete"`
	ExpirationProvisioning  schema.Date    `bson:"expiration_provisioning" json:"expiration_provisioning"`

	Json                string             `bson:"json" json:"json"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignatureFarmer     SigningSignature   `bson:"signature_farmer" json:"signature_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Metadata            string             `bson:"metadata" json:"metadata"`
	Result              Result             `bson:"result" json:"result"`
	WorkloadType        WorkloadTypeEnum   `bson:"workload_type" json:"workload_type"`
}

func (g *GatewaySubdomain) WorkloadID() int64 {
	return g.WorkloadId
}

func (g *GatewaySubdomain) GetWorkloadType() WorkloadTypeEnum {
	return g.WorkloadType
}

func (g *GatewaySubdomain) GetID() schema.ID {
	return g.ID
}

func (g *GatewaySubdomain) SetID(id schema.ID) {
	g.ID = id
}

func (g *GatewaySubdomain) GetJson() string {
	return g.Json
}

func (g *GatewaySubdomain) GetCustomerTid() int64 {
	return g.CustomerTid
}

func (g *GatewaySubdomain) GetCustomerSignature() string {
	return g.CustomerSignature
}

func (g *GatewaySubdomain) GetNextAction() NextActionEnum {
	return g.NextAction
}

func (g *GatewaySubdomain) SetNextAction(next NextActionEnum) {
	g.NextAction = next
}

func (g *GatewaySubdomain) GetSignaturesProvision() []SigningSignature {
	return g.SignaturesProvision
}

func (g *GatewaySubdomain) PushSignatureProvision(signature SigningSignature) {
	g.SignaturesProvision = append(g.SignaturesProvision, signature)
}

func (g *GatewaySubdomain) GetSignatureFarmer() SigningSignature {
	return g.SignatureFarmer
}

func (g *GatewaySubdomain) SetSignatureFarmer(signature SigningSignature) {
	g.SignatureFarmer = signature
}

func (g *GatewaySubdomain) GetSignaturesDelete() []SigningSignature {
	return g.SignaturesDelete
}

func (g *GatewaySubdomain) PushSignatureDelete(signature SigningSignature) {
	g.SignaturesDelete = append(g.SignaturesDelete, signature)
}

func (g *GatewaySubdomain) GetEpoch() schema.Date {
	return g.Epoch
}

func (g *GatewaySubdomain) GetMetadata() string {
	return g.Metadata
}

func (g *GatewaySubdomain) GetResult() Result {
	return g.Result
}

func (g *GatewaySubdomain) SetResult(result Result) {
	g.Result = result
}

func (g *GatewaySubdomain) GetDescription() string {
	return g.Description
}

func (g *GatewaySubdomain) GetCurrencies() []string {
	return g.Currencies
}

func (g *GatewaySubdomain) GetSigningRequestProvision() SigningRequest {
	return g.SigningRequestProvision
}

func (g *GatewaySubdomain) GetSigningRequestDelete() SigningRequest {
	return g.SigningRequestDelete
}

func (g *GatewaySubdomain) GetExpirationProvisioning() schema.Date {
	return g.ExpirationProvisioning
}

func (g *GatewaySubdomain) SetJson(json string) {
	g.Json = json
}

func (g *GatewaySubdomain) SetCustomerTid(tid int64) {
	g.CustomerTid = tid
}

func (g *GatewaySubdomain) SetCustomerSignature(signature string) {
	g.CustomerSignature = signature
}

func (g *GatewaySubdomain) SetEpoch(date schema.Date) {
	g.Epoch = date
}

func (g *GatewaySubdomain) SetMetadata(metadata string) {
	g.Metadata = metadata
}

func (g *GatewaySubdomain) SetDescription(description string) {
	g.Description = description
}

func (g *GatewaySubdomain) SetCurrencies(currencies []string) {
	g.Currencies = currencies
}

func (g *GatewaySubdomain) SetSigningRequestProvision(request SigningRequest) {
	g.SigningRequestProvision = request
}

func (g *GatewaySubdomain) SetSigningRequestDelete(request SigningRequest) {
	g.SigningRequestDelete = request
}

func (g *GatewaySubdomain) SetExpirationProvisioning(date schema.Date) {
	g.ExpirationProvisioning = date
}

type GatewayDelegate struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	PoolId     int64     `bson:"pool_id" json:"pool_id"`

	Description             string         `bson:"description" json:"description"`
	Currencies              []string       `bson:"currencies" json:"currencies"`
	SigningRequestProvision SigningRequest `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest `bson:"signing_request_delete" json:"signing_request_delete"`
	ExpirationProvisioning  schema.Date    `bson:"expiration_provisioning" json:"expiration_provisioning"`

	Json                string             `bson:"json" json:"json"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignatureFarmer     SigningSignature   `bson:"signature_farmer" json:"signature_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Metadata            string             `bson:"metadata" json:"metadata"`
	Result              Result             `bson:"result" json:"result"`
	WorkloadType        WorkloadTypeEnum   `bson:"workload_type" json:"workload_type"`
}

func (g *GatewayDelegate) WorkloadID() int64 {
	return g.WorkloadId
}

func (g *GatewayDelegate) GetWorkloadType() WorkloadTypeEnum {
	return g.WorkloadType
}

func (g *GatewayDelegate) GetID() schema.ID {
	return g.ID
}

func (g *GatewayDelegate) SetID(id schema.ID) {
	g.ID = id
}

func (g *GatewayDelegate) GetJson() string {
	return g.Json
}

func (g *GatewayDelegate) GetCustomerTid() int64 {
	return g.CustomerTid
}

func (g *GatewayDelegate) GetCustomerSignature() string {
	return g.CustomerSignature
}

func (g *GatewayDelegate) GetNextAction() NextActionEnum {
	return g.NextAction
}

func (g *GatewayDelegate) SetNextAction(next NextActionEnum) {
	g.NextAction = next
}

func (g *GatewayDelegate) GetSignaturesProvision() []SigningSignature {
	return g.SignaturesProvision
}

func (g *GatewayDelegate) PushSignatureProvision(signature SigningSignature) {
	g.SignaturesProvision = append(g.SignaturesProvision, signature)
}

func (g *GatewayDelegate) GetSignatureFarmer() SigningSignature {
	return g.SignatureFarmer
}

func (g *GatewayDelegate) SetSignatureFarmer(signature SigningSignature) {
	g.SignatureFarmer = signature
}

func (g *GatewayDelegate) GetSignaturesDelete() []SigningSignature {
	return g.SignaturesDelete
}

func (g *GatewayDelegate) PushSignatureDelete(signature SigningSignature) {
	g.SignaturesDelete = append(g.SignaturesDelete, signature)
}

func (g *GatewayDelegate) GetEpoch() schema.Date {
	return g.Epoch
}

func (g *GatewayDelegate) GetMetadata() string {
	return g.Metadata
}

func (g *GatewayDelegate) GetResult() Result {
	return g.Result
}

func (g *GatewayDelegate) SetResult(result Result) {
	g.Result = result
}

func (g *GatewayDelegate) GetDescription() string {
	return g.Description
}

func (g *GatewayDelegate) GetCurrencies() []string {
	return g.Currencies
}

func (g *GatewayDelegate) GetSigningRequestProvision() SigningRequest {
	return g.SigningRequestProvision
}

func (g *GatewayDelegate) GetSigningRequestDelete() SigningRequest {
	return g.SigningRequestDelete
}

func (g *GatewayDelegate) GetExpirationProvisioning() schema.Date {
	return g.ExpirationProvisioning
}

func (g *GatewayDelegate) SetJson(json string) {
	g.Json = json
}

func (g *GatewayDelegate) SetCustomerTid(tid int64) {
	g.CustomerTid = tid
}

func (g *GatewayDelegate) SetCustomerSignature(signature string) {
	g.CustomerSignature = signature
}

func (g *GatewayDelegate) SetEpoch(date schema.Date) {
	g.Epoch = date
}

func (g *GatewayDelegate) SetMetadata(metadata string) {
	g.Metadata = metadata
}

func (g *GatewayDelegate) SetDescription(description string) {
	g.Description = description
}

func (g *GatewayDelegate) SetCurrencies(currencies []string) {
	g.Currencies = currencies
}

func (g *GatewayDelegate) SetSigningRequestProvision(request SigningRequest) {
	g.SigningRequestProvision = request
}

func (g *GatewayDelegate) SetSigningRequestDelete(request SigningRequest) {
	g.SigningRequestDelete = request
}

func (g *GatewayDelegate) SetExpirationProvisioning(date schema.Date) {
	g.ExpirationProvisioning = date
}

type Gateway4To6 struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	PublicKey  string    `bson:"public_key" json:"public_key"`
	PoolId     int64     `bson:"pool_id" json:"pool_id"`

	Description             string         `bson:"description" json:"description"`
	Currencies              []string       `bson:"currencies" json:"currencies"`
	SigningRequestProvision SigningRequest `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest `bson:"signing_request_delete" json:"signing_request_delete"`
	ExpirationProvisioning  schema.Date    `bson:"expiration_provisioning" json:"expiration_provisioning"`

	Json                string             `bson:"json" json:"json"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignatureFarmer     SigningSignature   `bson:"signature_farmer" json:"signature_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Metadata            string             `bson:"metadata" json:"metadata"`
	Result              Result             `bson:"result" json:"result"`
	WorkloadType        WorkloadTypeEnum   `bson:"workload_type" json:"workload_type"`
}

func (g *Gateway4To6) WorkloadID() int64 {
	return g.WorkloadId
}

func (g *Gateway4To6) GetWorkloadType() WorkloadTypeEnum {
	return g.WorkloadType
}

func (g *Gateway4To6) GetID() schema.ID {
	return g.ID
}

func (g *Gateway4To6) SetID(id schema.ID) {
	g.ID = id
}

func (g *Gateway4To6) GetJson() string {
	return g.Json
}

func (g *Gateway4To6) GetCustomerTid() int64 {
	return g.CustomerTid
}

func (g *Gateway4To6) GetCustomerSignature() string {
	return g.CustomerSignature
}

func (g *Gateway4To6) GetNextAction() NextActionEnum {
	return g.NextAction
}

func (g *Gateway4To6) SetNextAction(next NextActionEnum) {
	g.NextAction = next
}

func (g *Gateway4To6) GetSignaturesProvision() []SigningSignature {
	return g.SignaturesProvision
}

func (g *Gateway4To6) PushSignatureProvision(signature SigningSignature) {
	g.SignaturesProvision = append(g.SignaturesProvision, signature)
}

func (g *Gateway4To6) GetSignatureFarmer() SigningSignature {
	return g.SignatureFarmer
}

func (g *Gateway4To6) SetSignatureFarmer(signature SigningSignature) {
	g.SignatureFarmer = signature
}

func (g *Gateway4To6) GetSignaturesDelete() []SigningSignature {
	return g.SignaturesDelete
}

func (g *Gateway4To6) PushSignatureDelete(signature SigningSignature) {
	g.SignaturesDelete = append(g.SignaturesDelete, signature)
}

func (g *Gateway4To6) GetEpoch() schema.Date {
	return g.Epoch
}

func (g *Gateway4To6) GetMetadata() string {
	return g.Metadata
}

func (g *Gateway4To6) GetResult() Result {
	return g.Result
}

func (g *Gateway4To6) SetResult(result Result) {
	g.Result = result
}

func (g *Gateway4To6) GetDescription() string {
	return g.Description
}

func (g *Gateway4To6) GetCurrencies() []string {
	return g.Currencies
}

func (g *Gateway4To6) GetSigningRequestProvision() SigningRequest {
	return g.SigningRequestProvision
}

func (g *Gateway4To6) GetSigningRequestDelete() SigningRequest {
	return g.SigningRequestDelete
}

func (g *Gateway4To6) GetExpirationProvisioning() schema.Date {
	return g.ExpirationProvisioning
}

func (g *Gateway4To6) SetJson(json string) {
	g.Json = json
}

func (g *Gateway4To6) SetCustomerTid(tid int64) {
	g.CustomerTid = tid
}

func (g *Gateway4To6) SetCustomerSignature(signature string) {
	g.CustomerSignature = signature
}

func (g *Gateway4To6) SetEpoch(date schema.Date) {
	g.Epoch = date
}

func (g *Gateway4To6) SetMetadata(metadata string) {
	g.Metadata = metadata
}

func (g *Gateway4To6) SetDescription(description string) {
	g.Description = description
}

func (g *Gateway4To6) SetCurrencies(currencies []string) {
	g.Currencies = currencies
}

func (g *Gateway4To6) SetSigningRequestProvision(request SigningRequest) {
	g.SigningRequestProvision = request
}

func (g *Gateway4To6) SetSigningRequestDelete(request SigningRequest) {
	g.SigningRequestDelete = request
}

func (g *Gateway4To6) SetExpirationProvisioning(date schema.Date) {
	g.ExpirationProvisioning = date
}
