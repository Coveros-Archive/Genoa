## FAQs

#### What is Genoa?
- A webhook based GitOps controller that can syncs desired helm (v3) release(s) from git and reconciles them in your kubernetes cluster.

#### What helm version is supported?
- Genoa only supports v3 helm releases.

#### What Git providers are supported?
- Genoa currently supports webhooks from github and gitlab.

#### Can I install helm charts from a private helm repository?
- Yes. This is part of configuring Genoa. You can add your custom repositories in Genoa values file during installation ( or upgrade ), and genoa will add those repositories.

#### How does it sync from git into cluster?
- Genoa uses webhook push events as a trigger to reconcile state from git into cluster. 

#### Can I manage existing helm (v3) releases with Genoa?
- Yes. Simply check in the current helm release state into github using the custom resource and it will start managing those existing releases.

#### What happens if I delete my desired state from git?
- Genoa will delete the corresponding helm release from your cluster.

