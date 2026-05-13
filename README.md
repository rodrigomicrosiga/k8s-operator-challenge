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

