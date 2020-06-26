package workloads

import schema "github.com/threefoldtech/tfexplorer/schema"

type (
	Workloader interface {
		WorkloadID() int64
		GetWorkloadType() WorkloadTypeEnum
		GetID() schema.ID
		SetID(id schema.ID)
		GetJson() string
		GetCustomerTid() int64
		GetCustomerSignature() string
		GetNextAction() NextActionEnum
		SetNextAction(next NextActionEnum)
		GetSignaturesProvision() []SigningSignature
		PushSignatureProvision(signature SigningSignature)
		GetSignatureFarmer() SigningSignature
		SetSignatureFarmer(signature SigningSignature)
		GetSignaturesDelete() []SigningSignature
		PushSignatureDelete(signature SigningSignature)
		GetEpoch() schema.Date
		GetMetadata() string
		GetResult() Result
		SetResult(result Result)
		GetDescription() string
		GetCurrencies() []string
		GetSigningRequestProvision() SigningRequest
		GetSigningRequestDelete() SigningRequest
		GetExpirationProvisioning() schema.Date
	}
)
