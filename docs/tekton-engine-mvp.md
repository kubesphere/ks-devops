# KubeSphere DevOps Tekton Engine MVP

这版实现不是用 Tekton 替换 Jenkins，而是在现有 KSE DevOps 控制面后增加第二个执行引擎。未声明执行引擎的流水线仍走 Jenkins；只有显式标记为 `tekton` 的流水线才由 Tekton 适配器处理。

## 设计边界

业界常见的多引擎 CI/CD 控制面会把“用户 API、权限、审计”与“实际执行器”分开。这样既能逐条迁移流水线，也能在出现兼容问题时快速切回旧引擎。本 MVP 沿用 KSE 的 `devops.kubesphere.io/v1alpha3` `Pipeline` 和 `PipelineRun` 作为用户 API，以原生 `tekton.dev/v1` `Pipeline` 和 `PipelineRun` 作为执行 API。

执行流程如下：

1. 用户或现有 KSE API 创建 KSE `PipelineRun`。
2. `Pipeline.spec.engine.type: tekton` 将该流水线路由到 Tekton；未配置时继续使用 Jenkins。
3. 适配器创建同名的原生 Tekton `PipelineRun`，传递字符串参数、ServiceAccount、超时和一个可选 Workspace。
4. 适配器把 Tekton 的开始时间、完成时间、`Succeeded` Condition 和阶段回写到 KSE `PipelineRun.status`。
5. KSE `spec.action: Stop` 会转换为 Tekton `spec.status: Cancelled`。

KSE Pipeline 与原生 Tekton Pipeline 当前是一层显式映射，而不是把 Jenkinsfile 自动翻译为 Tekton YAML。自动翻译 Groovy 通常不可可靠验证，建议在迁移期同时维护已验证的 Tekton Pipeline，再逐步把公共步骤沉淀为 Tekton Tasks。

## 引擎与注解契约

引擎使用强类型 Spec：

```yaml
spec:
  engine:
    type: tekton
```

CRD 只允许 `jenkins` 或 `tekton`。为兼容已经创建的 MVP 对象，缺少 `spec.engine` 时仍读取旧的 `devops.kubesphere.io/pipeline-engine` annotation，最终缺省为 Jenkins。新建 PipelineRun 会保存完整的 PipelineSpec 快照，因此 Pipeline 后续变更不会改变已经发起的运行。

| 注解 | 含义 |
| --- | --- |
| `devops.kubesphere.io/tekton-pipeline` | 原生 Tekton Pipeline 名称；缺省与 KSE Pipeline 同名 |
| `devops.kubesphere.io/tekton-service-account` | Tekton TaskRun 使用的 ServiceAccount |
| `devops.kubesphere.io/tekton-timeout` | Go duration 格式的 Pipeline 超时，例如 `15m` |
| `devops.kubesphere.io/tekton-workspace-name` | 原生 Pipeline 声明的 Workspace 名称 |
| `devops.kubesphere.io/tekton-workspace-claim` | Workspace 使用的 PVC 名称 |
| `devops.kubesphere.io/tekton-workspace-empty-dir` | 设为 `true` 时使用临时 Workspace；不能与 PVC 同时使用 |
| `devops.kubesphere.io/tekton-pipelinerun` | 适配器回写的原生 PipelineRun 名称 |

运行级 Tekton 配置注解优先于流水线级注解。通过 KSE API 发起运行时，允许的流水线级 Tekton 注解会被复制到新建的 KSE `PipelineRun`。

## 前置条件

- Kubernetes 集群已经安装 KSE DevOps CRD。
- 安装支持稳定 `tekton.dev/v1` API 的 Tekton Pipelines。先根据 KSE 所在 Kubernetes 版本核对 Tekton release 的兼容矩阵；下面的开发验证固定为 v1.13.0，不使用 `latest`：

  ```bash
  kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/previous/v1.13.0/release.yaml
  kubectl wait --namespace tekton-pipelines --for=condition=Available deployment/tekton-pipelines-controller --timeout=5m
  ```

- 构建并部署本分支的 `devops-controller` 镜像。沿用项目现有发布流程：

  ```bash
  export CONTROLLER_IMG=registry.example.com/kse/devops-controller:tekton-mvp
  make docker-build-controller CONTROLLER_IMG="${CONTROLLER_IMG}"
  docker push "${CONTROLLER_IMG}"
  make deploy CONTROLLER_IMG="${CONTROLLER_IMG}"
  ```

现有 KSE 环境如果由 Helm 或扩展组件管理，应在对应 values 中替换 controller 镜像，不要并行部署两个 ks-devops controller-manager。`config/rbac/role.yaml` 已包含对 `tekton.dev/pipelineruns` 的最小权限。

## 单元测试

仓库要求 Go 1.23 或更新版本。运行定向测试：

```bash
go test ./pkg/pipelineengine ./controllers/tekton ./pkg/kapis/devops/v1alpha3/pipelinerun
```

测试覆盖默认 Jenkins 路由、Spec 优先级、PipelineSpec 快照、兼容注解传播、Tekton 对象转换、创建原生 PipelineRun、成功状态回写和取消传播。

## 集群冒烟测试

确认新 controller 镜像和 Tekton 已就绪后运行：

```bash
./hack/smoke-tekton-mvp.sh
```

脚本只创建测试资源，不执行清理操作。它会创建一个带随机后缀的 KSE `PipelineRun`，等待同名 Tekton `PipelineRun` 成功，再检查 KSE 状态已经变成 `Succeeded`。

也可以手动执行：

```bash
kubectl apply -k config/samples/tekton
kubectl create -f config/samples/tekton/kse-pipelinerun.yaml
kubectl get pipelineruns.devops.kubesphere.io -n kse-tekton-demo
kubectl get pipelineruns.tekton.dev -n kse-tekton-demo
```

停止一个仍在运行的 KSE PipelineRun：

```bash
kubectl patch pipelinerun.devops.kubesphere.io RUN_NAME -n kse-tekton-demo --type=merge -p '{"spec":{"action":"Stop"}}'
```

## 镜像安全门禁

`secure-image-pipeline.yaml` 提供了一个 Trivy 阻断门禁：扫描发现可修复的 `CRITICAL` 漏洞时 Task 以非零状态结束。先部署示例定义，再创建一次扫描：

```bash
kubectl create -f config/samples/tekton/secure-image-kse-pipelinerun.yaml
```

示例为了可读性固定为 `aquasec/trivy:0.70.0`。生产环境必须把构建器、扫描器和业务镜像都锁定到经过验证的 digest，并为扫描数据库配置可信镜像源或内部缓存。建议把 SBOM 生成、镜像签名和准入策略作为后续 Tasks，而不是把所有安全逻辑写进控制器。

## 当前限制与下一步

- KSE 只镜像 PipelineRun 总体状态；TaskRun 阶段、日志和结果仍从 Tekton API 或 `tkn` 查看。
- 仅支持字符串参数和一个注解式 Workspace；数组/对象参数及多 Workspace 需要扩展 KSE API 或增加独立配置 CRD。
- SCM 触发、多分支发现、定时触发仍是 Jenkins 能力，尚未映射到 Tekton Triggers。
- AI Code Review 需要确定代码托管平台、模型服务、数据出境与凭据策略。推荐实现成独立 Tekton Task，在拉取代码后执行，把结果回写 PR/MR；不要把模型调用耦合进 PipelineRun 控制器。
- 下一阶段应增加 Tekton Results 持久化、TaskRun/日志聚合、Trigger 适配、供应链签名与策略准入，并补充真实集群 e2e 测试。
