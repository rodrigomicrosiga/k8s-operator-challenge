# k8s-operator-challenge
K8S-OPERATOR-CHALLENGE

DETALHE DE CADA ETAPA:

- Kubebuilder 
- Tilt
- Build da imagem
- Carregar no cluster
- Instalar CRDs/operator 
- Testar um CR

ETAPA 1

1. Criação do repo

2. Criação do projeto
kubebuilder init --domain teste.com --repo github/rodrigomicrosiga/k8s-operator-challenge
kubebuilder create api --group apps --version v1 --kind OperatorChallenge
    NESSA ETAPA SÃO CRIADOS OS RECURSOS E CONTROLADORES

3. Criação dos artefatos
make generate
make manifests

4. Implementar o Reconcile
internal/operatorchallenge_controller.go
internal/operatorchallenge_controller_test.go
internal/suite_test.go

RESUMO DA ETAPA 1: 
Repositório Go com api/, controllers/, config/ e utilização de comandos make para gerar manifests.

Ao realizar as alterações nos códigos abaixo: 

- operatorchallenge_types.go
- operatorchallenge_controller.go

Foi necessário realizar os processos make generate e make manifests, porém obtive o problema abaixo:

"/home/rodrigobsantos@sp01.local/code/k8s-operator-challenge/bin/controller-gen" object:headerFile="hack/boilerplate.go.txt",year=2026 paths="./..."
cmd/main.go:39:2: found packages controllers (operatorchallenge_controller.go) and controller (suite_test.go) in /home/rodrigobsantos@sp01.local/code/k8s-operator-challenge/internal/controller
Error: not all generators ran successfully
run `controller-gen object:headerFile=hack/boilerplate.go.txt,year=2026 paths=./... -w` to see all available markers, or `controller-gen object:headerFile=hack/boilerplate.go.txt,year=2026 paths=./... -h` for usage
make: *** [Makefile:52: generate] Erro 1

Procurando entender mais o erro, foi possivel analisar que o controller_test e suite_test estavam importando packges diferentes e isso provocava o erro.

No momento de geração do go build ./... várias inconsistências foram apresentadas (essa etapa exige mais estudo e praticar)

Resumo simples de correções:

- logger: agora é criado logo no início do Reconcile com logger := log.FromContext(ctx) e usado em todo o método.
- challengev1: o código usa o alias correto para o pacote api/v1 (challengev1), removendo a referência incorreta appsv1alpha1.
- svcKey: variável removida porque não estava sendo usada, em vez disso, quando necessário, eu construo types.NamespacedName inline nas chamadas de log.
- Tratamento de erros e logs: adicionei logs mais descritivos para facilitar debug. (necessário revisão)

Após os pontos acima foi executado:
gofmt -w internal/controller/operatorchallenge_controller.go
go build ./...
make generate
make manifests
go run ./main.go