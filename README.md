# k8s-automation

## 조건

karpenter 와 prometheus mimir proxy 를 사용중일 때 아래와 동일하게 사용 가능하며 

prometheus 만을 사용할 경우 config/prometheus_client.go 파일 수정 후, cmd/root.go 에서 prometheusOrgID 부분을 제거하여 사용하면 됩니다.

## Start

로컬에서 사용 시 아래를 순차적으로 적용하면 됩니다.

kube context 를 변경하여 EKS 워크로드 노드를 드레인하고 싶은 클러스터의 context 로 위치시킵니다.

K9S 혹은 Open Lens 를 사용하여 deployment 접근, mimir 혹은 prometheus-server 를 검색 후 8080:8080 으로 포트포워딩

포트포워딩 설정 이후 아래 스크립트를 실행합니다.

```sh
go mod tidy
```

node dry run
```sh
go run main.go dry-run \
  --percentage 30 \
  --prometheus-address "http://localhost:8080/prometheus" \
  --prometheus-org-id "organzation-dev" \
  --drain-node-label-key "my-workload-label-key" \
  --drain-node-label-value "my-workload-label1, my-workload-label2" \
  --slack-webhook-url "https://hooks.slack.com/services/xxx" \
  --kube-config "local" 
```

node drain
```sh
go run main.go drain \
  --percentage 30 \
  --prometheus-address "http://localhost:8080/prometheus" \
  --prometheus-org-id "organzation-dev" \
  --drain-node-label-key "my-workload-label-key" \
  --drain-node-label-value "my-workload-label1, my-workload-label2" \
  --slack-webhook-url "https://hooks.slack.com/services/xxx" \
  --kube-config "local"
```

disk usage 확인
```sh
go run main.go usage disk \
  --percentage 30 \
  --prometheus-address "http://localhost:8080/prometheus" \
  --prometheus-org-id "organzation-dev" \
  --drain-node-label-key "my-workload-label-key" \
  --drain-node-label-value "my-workload-label1, my-workload-label2" \
  --slack-webhook-url "https://hooks.slack.com/services/xxx" \
  --kube-config "local"
```

memory usage 확인
```sh
go run main.go usage memory \
  --percentage 30 \
  --prometheus-address "http://localhost:8080/prometheus" \
  --prometheus-org-id "organzation-dev" \
  --drain-node-label-key "my-workload-label-key" \
  --drain-node-label-value "my-workload-label1, my-workload-label2" \
  --slack-webhook-url "https://hooks.slack.com/services/xxx" \
  --kube-config "local"
```

## 워크로드 노드 정리 플로우

### 정리 대상 노드 식별
메모리 사용률이 N% 미만인 모든 노드를 리스트업 합니다.

### 우선순위 설정
메모리 사용률이 가장 낮은 노드부터 순서를 매깁니다.
노드의 age(생성 시간)나 다른 메트릭을 추가로 고려할 수 있습니다.

### 단계적 cordon 적용
1. 리스트 업 된 노드들에 cordon 을 적용합니다.
2. Cordon 된 노드들을 순서대로 Drain 한 뒤, N분 동안 대기합니다.
3. 문제가 없다면 다음 노드를 드레인합니다. 문제가 발생한다면 프로세스를 일시 중지하고 상황을 평가합니다.

### 정리 프로세스 시작
Cordon 된 노드들의 파드들에 대해 graceful shutdown 프로세스를 시작합니다.
한 노드의 정리가 완료되면 다음 노드로 넘어갑니다.(파드들이 다른 노드에 안정적으로 갈 수 있도록 5분 텀을 두도록 합니다.)

### 모니터링 및 조정
전체 프로세스 동안 클러스터의 상태를 지속적으로 모니터링하여 적절한 메모리 퍼센테이지를 찾습니다.
