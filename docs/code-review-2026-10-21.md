# 源码优化审查（2026-10-21）

## 范围与结论

本轮做了全仓文件清单、依赖、构建与高风险模式扫描，重点阅读后台、主题、渲染器的自有 JavaScript、编辑器调用链，以及相关 Go 服务实现。仓库有 642 个 Git 跟踪文件。没有逐行审阅所有 Go 文件、第三方压缩脚本和生成的 protobuf 代码，也没有进行完整浏览器回归或安全审计，因此这是一轮有验证依据的重点审查，不代表全部源码已经穷尽审阅。

项目已经使用 async/await、可选链、空值合并、Web Components、IntersectionObserver、requestAnimationFrame 等现代写法。主要收益来自异步正确性、资源释放、模块边界和静态检查；保持 Go 模板与原生 JS 很合理。

本轮仅添加此审查记录，没有修改业务代码。开始时已有的 gateway/main.go 改动未触碰。

## 应优先修复的功能问题

### 1. 预览响应乱序，以及旧保存响应清除新内容的未保存标记

后续处理（2026-10-22）：已增加预览请求编号与输入版本检查；自动保存串行执行，每次发送时使用上一笔保存返回的 modified。旧响应不再覆盖预览或清除后续输入的 dirty 状态，旧错误也不会覆盖新预览；单次失败不阻塞后续保存。删除原先无效的 `_previewCallbackReturned` 标记。新增 `tests/js/editor.test.cjs`，通过 10 项回归测试（`node --test tests/js/editor.test.cjs`），覆盖 textarea、TinyMDE、防抖期间输入、乱序成功/失败、保存排队和失败恢复。另通过 JavaScript 语法检查；尚未做真实浏览器回归。以下保留最初的问题说明。

位置：[editor.js](/Users/tao/code/taoblog/admin/statics/editor.js:1879)、[sourceChanged](/Users/tao/code/taoblog/admin/statics/editor.js:2120)。

防抖只推迟发送，不能保证已经发出的请求按顺序返回。当前每个响应都会 setPreview；自动保存响应也无条件 `_setContentChanged(false)`。`_previewCallbackReturned` 只被赋成 true，没有形成有效的请求互斥。

例如：A 请求发出后继续输入 B，B 先返回，A 后返回，会把预览改回 A。如果 A 保存响应到来时 B 尚未保存，还会提前清除未保存状态。

建议：预览使用请求序号，只接受最新响应；保存记录内容版本，只清除与该响应对应版本的 dirty 状态。自动保存写请求应排队或合并为“当前请求完成后保存最新版本”，继续保留服务端 modified 冲突检查。AbortController 可以减少预览浪费，但取消客户端请求不保证服务端写入没有发生。

验证：通过 Node VM 调用实际 PostFormUI.updatePreview，按 new、old 顺序返回，最终显示 old，已复现预览乱序。

### 2. 图片解密中的异步错误没有被正确等待或捕获

后续处理（2026-10-22）：已用两层 Promise.all 等待全部 source 及其候选地址；候选解密使用 await，使失败能被 catch 捕获并回退原 URL，同时保留描述符和顺序。获取元信息与加密文件时先检查 HTTP 状态，主文件错误保留 cause。通过 8 项本地回归测试，使用真实 AES-GCM 验证解密，并覆盖延迟候选、部分失败、HTTP 错误和认证失败；测试文件仅保留本地，不纳入提交。另通过 JavaScript 语法检查，尚未做真实浏览器回归。以下保留最初的问题说明。

位置：[encrypted/script.js](/Users/tao/code/taoblog/service/modules/renderers/encrypted/script.js:49)。

外层 `sources.forEach(async ...)` 不会被 decryptFile 等待。内层 `try { promises.push(decrypt(set.url)); }` 也捕获不到 decrypt 返回的 Promise 的异步拒绝，预期的原 URL 回退不会执行。

建议：使用 `await Promise.all(Array.from(sources, async source => ...))`；每个地址用 `try { return await decrypt(url); } catch { return url; }`，或按产品需求整体报错。fetch 后先检查 HTTP 状态，错误保留 cause。Blob URL 的释放见第 10 项。

验证：模拟一个 source 下载失败，调用实际 decryptFile 后，它先正常返回，随后产生未处理的“图片解码失败。”拒绝。测试模拟了网络与加密 API，不是浏览器端加密端到端测试。

参考：[MDN forEach 的异步行为](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/forEach)。

### 3. 文件路径使用 encodeURI，特殊字符改变了请求含义

后续处理（2026-10-22）：已新增 FilesManager._fileURL，按 `/` 分段进行 encodeURIComponent 编码，读取、删除和修改注释统一使用该地址构造方法，`:caption` 后缀在文件路径编码后追加。通过 27 项本地 JavaScript 测试，并用本地 Go 路由测试验证同样 27 组 GET/DELETE/PATCH 路径匹配；覆盖目录、中文、emoji、`#`、`?`、`%`、`+`、反斜杠、字面量 `%2F` 和文件名中的 `:caption`。当前网关默认解码行为可正确还原这些路径，无需改后端。测试文件仅保留本地，不纳入提交。尚未做真实浏览器回归。以下保留最初的问题说明。

位置：[FilesManager](/Users/tao/code/taoblog/admin/statics/editor.js:27)，同文件第 35、50 行也有。

encodeURI 会保留 `#`、`?`，适合处理整个 URI，不适合直接编码任意文件路径。文件名 `a#b?.png` 拼接后，服务器收到的 pathname 只有 `/v3/posts/1/files/a`。

建议：支持目录时按 `/` 分段后分别 encodeURIComponent，再用 `/` 拼回；并验证网关对编码路径的解析。get、delete、updateCaption 共用一个路径构造函数。

验证：Node URL 解析已复现 pathname 截断。

### 4. 手写 Promise 缺少失败出口

后续处理（2026-10-22）：FileReader 已补齐 error/abort 拒绝出口，图标和头像调用处捕获读取失败，头像加载状态通过 finally 恢复。XHR 上传已捕获 JSON 解析异常，abort/error/timeout/异常响应状态和 HTTP 错误均返回错误对象；上传进度忽略无法计算或总量为零的事件。decodeAPIError 对 null、空响应和非字符串 message 使用 HTTP 状态回退，并保留原响应为 cause。通过 25 项新增本地测试，覆盖读取成功/失败/取消、损坏上传响应、重试入口、HTTP 错误和头像加载恢复；测试文件仅保留本地，不纳入提交。未新增固定上传超时，timeout 处理用于收到该事件的情况。尚未做真实浏览器回归。以下保留最初的问题说明。

位置：[readFileAsDataURL](/Users/tao/code/taoblog/admin/dynamic/utils.js:32)、[XHR 上传](/Users/tao/code/taoblog/admin/statics/editor.js:92)。

FileReader 只处理 onload，没有 onerror、onabort，读取失败后 Promise 永远 pending。XHR 的 load 回调直接 JSON.parse；响应格式异常时，回调抛错不会自动 reject 外面的 Promise。

建议：FileReader 补齐 reject 出口；XHR 可设置 `responseType = 'json'` 并验证响应，或把 JSON.parse 放入 try/catch 后 reject。统一使用 Error 对象。XHR 用于上传进度有实际用途，可以保留。

验证：模拟 FileReader 触发 onerror，实际函数返回的 Promise 仍在等待。另测得 decodeAPIError 遇到合法 JSON `null` 会因访问 `.message` 抛 TypeError，应加类型判断和默认错误信息。

### 5. 临时服务端故障被当成登录失效

后续处理（2026-10-22）：已区分 401 与其他失败状态；只有 401 或成功且有效的用户编号比较确认不匹配时才调用登录，网络、服务端、JSON 或用户编号异常仅记录检查失败。检查请求禁用缓存；激活事件改为真正防抖，保留请求与登录期间的互斥，并忽略页面隐藏、离开后台或用户 Cookie 已改变后的过时结果。Cookie 按第一个等号分割，逐条隔离编码异常，非法用户编号归零。通过 44 项新增本地测试，覆盖错误状态、401、异常响应、防抖、互斥、失效结果和 Cookie 边界；测试文件仅保留本地，不纳入提交。尚未做真实浏览器与 WebAuthn 设备回归。以下保留最初的问题说明。

位置：[cookies.js](/Users/tao/code/taoblog/admin/dynamic/cookies.js:44)。

响应非 ok 时把 serverUserID 设成 0，随后与已登录用户 ID 比较。因此 500、502 等错误也会触发 WebAuthn 登录。

建议：明确 401 才走失效流程；其他非成功响应作为检查失败处理；只有成功解析后才比较用户。激活事件的延时改为真正防抖，避免排队多个检查。另外 Cookie 解析应只在第一个等号处分割，并隔离单个无效百分号编码，避免无关 Cookie 破坏登录脚本。

### 6. Tldraw 保存异常后按钮无法恢复

后续处理（2026-10-22）：保存状态已收敛为 idle/saving/saved，保存流程通过 try/catch/finally 展示错误并恢复按钮；增加同步互斥防止快速重复点击并发保存。成功提示计时器仅在状态变化时重建，并清理重复的 display 字段。现有 ESLint 与生产构建通过，已同步 admin/statics/tldraw.min.js，生成 CSS 无变化。通过 10 项本地 JSX 组件逻辑测试，使用轻量 Hook/元素模拟覆盖导出、快照、上传异常后的重试，保存返回 false、空画布、重复点击和提示计时器；测试文件仅保留本地，不纳入提交。尚未做真实浏览器回归。以下保留最初的问题说明。

位置：[App.jsx](/Users/tao/code/taoblog/admin/tldraw/src/App.jsx:81)。

`await save()` 抛错时，后续两个状态重置都被跳过，按钮会继续 disabled。建议用 try/catch/finally 恢复保存状态并展示错误。保存状态可收敛为 idle/saving/saved/error，减少两个布尔值互相矛盾的机会。

现有 ESLint 还实际报告第 77 行重复定义 `display`，应去掉重复字段。第 48 行 useEffect 可声明 `[showCheckMark]`，避免无关重渲染重置计时器。

### 7. DrawIO 消息只检查 origin，没有检查具体 iframe

后续处理（2026-10-22）：DrawioEditor 已保存 iframe 引用，同时检查消息 origin 和 source，并校验 JSON 消息及导出字段，忽略无效 Base64；组件重新连接时复用原 iframe。草稿页同时检查同源、编号与对应 iframe 的 contentWindow，校验 dirty 布尔值及 saved 标题/时间字段，忽略未知或已移除窗口的消息。通过 38 项新增本地模拟测试，覆盖两个 DrawIO 实例的配置、加载、保存、关闭隔离，以及草稿消息伪造和字段异常。全部 162 项本地回归测试及语法检查通过；测试文件仅保留本地，不纳入提交。尚未使用真实浏览器与 DrawIO 服务回归。以下保留最初的问题说明。

位置：[DrawioEditor](/Users/tao/code/taoblog/admin/statics/editor.js:1051)。

同时打开两个 DrawIO 文件时，两个实例都监听 window message；二者 origin 相同，因此一个 iframe 的 init/export/exit 消息会被多个实例处理，存在保存到错误文件或关闭多个标签的风险。

建议：保存创建的 frame（当前 this._iframe 仍为 null），同时检查 `e.source === this._iframe.contentWindow`。草稿页 [drafts.js](/Users/tao/code/taoblog/admin/statics/drafts.js:189) 也应校验 origin、来源 iframe 和消息字段，防止无关消息改变 dirty 状态。

参考：[MDN postMessage](https://developer.mozilla.org/en-US/docs/Web/API/Window/postMessage)。此处根据事件注册和处理逻辑确认，未用两个真实 iframe 做浏览器回归。

### 8. 搜索引擎启动超过 30 秒后会卡住后台刷新

后续处理（2026-10-22）：sync.Once 仅启动后台任务，各请求独立等待就绪、30 秒超时或请求/服务取消；通过关闭 channel 广播初始化结果，发送方不再依赖仍有接收者。初始化失败记录并返回错误，不再使用 log.Fatalln 结束整站；首次索引失败也会被传递，不会报告初始化成功。失败结果保留至服务重启，暂未增加自动重试策略。检查退出路径时另发现 Engine.Close 持写锁调用 getWriter 再取读锁会死锁，已改为直接关闭已有 writer，支持重复关闭，关闭后拒绝搜索/索引。9 项本地单测、竞态检查、现有 2 项搜索引擎测试，以及 1 项真实本地服务端到端测试（8 个并发首次搜索）通过；超时路径使用 testing/synctest 虚拟时钟验证。Go 测试使用 -vet=off，跳过问题 9 已记录的既有复制锁值诊断，未宣称 vet 通过。测试文件使用 *_local_test.go，仅保留本地，不纳入提交。以下保留最初的问题说明。

位置：[search.go](/Users/tao/code/taoblog/service/search.go:20)、[通知发送](/Users/tao/code/taoblog/service/search.go:72)。

ch 是无缓冲 channel。调用方等 30 次后退出；若初始化更慢，后台执行到 `ch <- struct{}{}` 时已没有接收者，将永久阻塞，无法进入定期索引循环或处理 ctx.Done。引擎指针此前已经发布，所以表现可能是可以搜索，但后续修改不再被索引。

建议：一次性就绪通知使用 close(ch)，把启动动作与请求等待分开；请求等待同时处理 ctx.Done，避免长时间占住 sync.Once.Do。初始化失败返回或记录错误，避免用 log.Fatalln 直接结束整站进程。

此项为代码路径分析，未等待真实搜索初始化超过 30 秒。

### 9. Service 事务通过复制整个对象实现，包含锁和原子状态

位置：[TxCall](/Users/tao/code/taoblog/service/main.go:332)。

`txs := *s` 复制了 sync.Map、sync.Mutex、sync.Once、atomic 等不可随意复制的运行时状态。建议把事务使用的数据访问部分做成独立、可安全复制的小对象，共享服务状态保留同一实例。

定向 go vet 实际发现三处复制锁值：

- service/main.go:337：复制 Service。
- service/post.go:480：range 按值复制含 protobuf 状态的 Post，可先改为按索引取得原元素地址。
- service/post.go:766：把含该状态的 Post 按值传给 Model；需结合 ORM 约定改用指针或专门数据库模型。

这些是静态诊断，不等于已经复现死锁。尤其第 766 行局部值当时可能尚未使用内部锁，但数据模型的可复制性仍值得清理。

## 值得做的结构与写法改进

### 10. 用加载事件与清理方法替代无限轮询

位置：[TableEditor](/Users/tao/code/taoblog/admin/statics/editor.js:951)、[TldrawEditor](/Users/tao/code/taoblog/admin/statics/editor.js:1148)、[React 入口](/Users/tao/code/taoblog/admin/tldraw/src/main.jsx:5)。

当前通过 setInterval 等待全局函数；脚本失败时一直轮询，多实例可能重复插入脚本，标签关闭后仍可能继续初始化。

建议：缓存一个共享加载 Promise，监听 script load/error；组件断开时停止后续挂载，重连时防止重复初始化。React 入口返回 root 或卸载函数，在标签销毁时 root.unmount()，清理组件订阅。

解密创建的 Blob URL 也应有统一所有权和释放时机；灯箱会复用原始 URL，不能图片刚 load 完就直接 revoke，要等所有使用者结束。

参考：[React createRoot 与 unmount](https://react.dev/reference/react-dom/client/createRoot)。

### 11. 评论 HTML 不应藏在函数注释再切字符串

后续处理（2026-10-22）：已改为原生模板字符串，修复 ios-toggle 开始标签缺少 `>`，并显式补齐评论根容器的结束标签；保留现有内联事件与初始化方式，事件委托作为后续可选改进。通过 JavaScript 语法检查、与旧模板的内容比对、移除注释并压缩后的模板完整性验证，以及 Go HTML 解析结构检查。本次仅做临时验证，没有新增或提交测试文件；尚未做真实浏览器回归。以下保留最初的问题说明。

位置：[comment.js](/Users/tao/code/taoblog/theme/share/comments/comment.js:1)。

`function(){/* HTML */}.toString().slice(14,-3)` 是模板字符串普及前的变通方式，依赖固定切片和注释保留。可以直接换成模板字符串或 HTMLTemplateElement。当前主题只是拼接脚本，尚不能断言现在线上会因压缩丢失模板，但此写法会妨碍后续构建优化。

该模板第 37 行 ios-toggle 的开始标签还缺少 `>`，应一起纠正。

评论点击处理可改为容器事件委托与 data-action，减少内联 onclick 和 window.comment 耦合；完成后再考虑迁移模块，避免一次改变过多加载规则。

### 12. 统一交互事件与表单提交

后续处理（2026-10-22）：评论拖动与音频进度条已改用 Pointer Events、指针捕获及 pointerup/pointercancel/lostpointercapture 收尾，不再覆盖 document 的全局鼠标处理器；交互区域设置 touch-action，忽略其他指针和非主按钮。音频拖动保留播放/暂停状态，结束交互或 change 时提交最终位置并清除旧计时器，播放失败被捕获。评论和文章统一监听 form submit，校验表单并阻止重复提交；文章快捷键使用 requestSubmit，工具栏等非保存按钮显式设置 type=button。评论提交通过 finally 恢复按钮。通过 20 项新增本地交互模拟测试、全部 182 项本地 JS 回归、文章表单 HTML 结构检查、音频渲染器既有 Go 测试及语法检查。测试仅保留本地，不纳入提交；尚未做真实浏览器的触摸、原生 range 与键盘回归。以下保留最初的问题说明。

位置：[评论拖动](/Users/tao/code/taoblog/theme/share/comments/comment.js:726)、[音频进度](/Users/tao/code/taoblog/service/modules/renderers/audio_player/script.js:28)。

旧的 document.onmousemove/onmouseup 会占用全局单一事件槽；音频仅处理鼠标事件，拖出控件释放时也可能漏掉结束。可改为 Pointer Events，使用 pointer capture 和 pointercancel 做完整收尾。

评论和文章表单目前主要截获提交按钮 click，可逐步迁到 form submit，统一键盘提交和校验。`let self = this`、`function(){}.bind(this)` 可在需要词法 this 时改成箭头函数，但这是可读性改进，不值得全仓机械替换。

### 13. 评论建树反复全量扫描

位置：[CommentListUI.insert](/Users/tao/code/taoblog/theme/share/comments/comment.js:357)。

每处理一个评论又 comments.filter 扫描全部评论，建树最坏 O(n²)。先按 parent 建 Map，分组排序，再遍历，建组可以降到 O(n)。小评论区收益有限，优先级低于保存和解密问题。

### 14. 日期和坐标的语义应明确

位置：[日期值](/Users/tao/code/taoblog/theme/share/components/datetime-picker.js:67)、[时区选择](/Users/tao/code/taoblog/theme/share/components/datetime-picker.js:133)、[坐标回填](/Users/tao/code/taoblog/admin/statics/editor.js:2005)。

日期组件使用浏览器本地 new Date 生成时间戳，选择其他时区只修改 timezone 字段。需明确设计：如果时区表示“用户输入的墙上时间所属时区”，当前换算不正确；如果只是给同一时刻附加来源信息，应把这个含义体现到界面。此项需要先确定产品语义，不能简单替换一个 API。

坐标 setter 只在经纬度均大于 0 时回填，会漏掉南纬、西经和零度位置；getter 的 parseFloat 还会接受 `12abc`。建议显式检查 Number.isFinite 与合法范围，是否为空与数值正负分别处理。

### 15. Go HTTP 错误路径与比较器

位置：[高德请求](/Users/tao/code/taoblog/modules/geo/resolver.go:56)。

非 200 响应先 panic，再注册 defer rsp.Body.Close，错误路径不会关闭 Body。应在成功取得响应后立刻 defer 关闭。距离比较器 `int(a.Distance-b.Distance)` 会把小于 1 的差截为 0，建议用 cmp.Compare 比较浮点数。

[OpenURLAsFile](/Users/tao/code/taoblog/modules/utils/http.go:110) 没有 context 参数，也使用无总超时的默认客户端；[PlantUML](/Users/tao/code/taoblog/service/modules/renderers/plantuml/plantuml.go:25) 支持 context，但部分调用点传 Background。建议优先给外网调用建立明确超时，并区分请求生命周期与后台任务生命周期，避免机械替换全部 Background。

## 工具与实施顺序

1. 先修预览/保存版本控制、图片解密等待、URL 编码和 Promise 失败出口。
2. 再修 iframe 消息隔离、搜索启动通知、Service 复制锁和 HTTP 错误路径。
3. 给自有 JS 增加全仓 ESLint 配置；现有配置仅在 admin/tldraw。先启用 no-dupe-keys、no-unreachable、no-unsafe-finally 等正确性规则，再逐渐收紧风格规则。
4. 基于现有 JSDoc，选 API 层和编辑器逐步开启 checkJs。当前脚本是 Go 动态拼接，跨文件全局符号需要声明，不能直接把所有文件设成 ES module。
5. 拆分 2276 行的 editor.js：文件 API、文件列表、嵌入编辑器、文章表单各自成文件；保留对外行为，再按需要建立模块打包入口。

不建议为了“现代化”整体重写 React/Vue、全量迁 TypeScript、替换所有 forEach、删除所有 return await，或把有撤销兼容性用途的 execCommand 直接删掉。const、严格相等、对象简写等可随局部修改整理；其中涉及接口返回类型时先确认类型再改。

实际执行的检查：

- admin/tldraw 下执行现有 ESLint：1 个错误，重复 display。
- `GOCACHE=/private/tmp/taoblog-review-go-cache GOTOOLCHAIN=local GOPROXY=off go vet ./service ./modules/geo`：3 条复制锁值诊断。
- Node VM 最小验证：实际函数的预览乱序、解密未处理拒绝、FileReader 失败不结束、decodeAPIError(null)，以及 URL 编码示例。
- 没有运行完整 Go 测试集、浏览器回归、性能基准或依赖漏洞扫描；不对这些方面宣称验证通过。
