# 🔀 ccswitch

由 [Kyle Redelinghuys](https://ksred.com) 开发

一个友好的 CLI 工具，用于管理多个 git worktree（工作树），非常适合同时处理不同功能、实验或 Claude Code 会话，而无需为上下文切换而烦恼。

## 🎯 这是什么？

`ccswitch` 帮助您通过简洁直观的界面创建和管理 git worktree。每个 worktree 都有自己独立的目录，让您可以同时在多个功能上工作，无需 stash 变更或就地切换分支。

## ✨ 特性

- **🚀 快速创建会话** - 描述您正在处理的内容，即可立即获得分支和 worktree
- **📋 交互式会话列表** - 通过简洁的 TUI 界面查看所有活跃的工作会话
- **🧹 智能清理** - 完成后删除 worktree，可选是否同时删除分支
- **🗑️ 批量清理** - 使用 `cleanup --all` 一次性删除所有 worktree（适合大扫除！）
- **🐚 Shell 集成** - 自动 `cd` 进入新的 worktree（无需复制粘贴路径！）
- **🎨 美观输出** - 彩色消息和整洁的格式

## 📦 安装

### 使用 Make
```bash
# 克隆仓库
git clone https://github.com/ksred/ccswitch.git
cd ccswitch

# 构建并安装
make install

# 将 shell 集成添加到您的 .bashrc 或 .zshrc
cat bash.txt >> ~/.bashrc  # 或 ~/.zshrc
source ~/.bashrc           # 或 ~/.zshrc
```

### 手动安装
```bash
# 构建二进制文件
go build -o ccswitch .

# 移动到您的 PATH
sudo mv ccswitch /usr/local/bin/

# 添加 shell 包装器
source bash.txt
```

## 🚀 使用方法

### 创建新的工作会话
```bash
ccswitch
# 🚀 What are you working on? Fix authentication bug
# ✓ Created session: feature/fix-authentication-bug
#   Branch: feature/fix-authentication-bug
#   Path: /home/user/project/../fix-authentication-bug
#
# Automatically switches to the new directory!
```

### 列出活跃会话
```bash
ccswitch list
# 显示所有 worktree 的交互式列表
# 使用方向键导航，回车选择，q 退出
```

### 在会话之间切换
```bash
ccswitch switch
# 交互式选择要切换到的会话

ccswitch switch fix-auth-bug
# 直接切换到指定会话
# 自动切换到会话目录！
```

### 完成后清理
```bash
ccswitch cleanup
# 交互式选择会话，或：

ccswitch cleanup fix-authentication-bug
# Delete branch feature/fix-authentication-bug? (y/N): y
# ✓ Removed session and branch: fix-authentication-bug

# 批量清理 - 一次性删除所有 worktree！
ccswitch cleanup --all
# ⚠️  You are about to remove the following worktrees:
#   • feature-1 (feature/feature-1)
#   • feature-2 (feature/feature-2)
#   • bugfix-1 (feature/bugfix-1)
# Press Enter to continue or Ctrl+C to cancel...
# Delete associated branches as well? (y/N): y
# ✓ Successfully removed: feature-1
# ✓ Successfully removed: feature-2
# ✓ Successfully removed: bugfix-1
# ✅ All 3 worktrees removed successfully!
# ✓ Switched to main branch
```

## 🛠️ 开发

### 快速开始
```bash
# 直接运行
make run

# 运行测试
make test

# 查看所有命令
make help
```

### 测试
```bash
# 仅单元测试（快速，无需 git）
make test-unit

# 包含集成的所有测试
make test

# 在 Docker 中运行测试（清洁环境）
make test-docker

# 生成覆盖率报告
make coverage
```

### 项目结构
```
ccswitch/
├── main.go              # 主应用程序代码
├── bash.txt             # Shell 集成包装器
├── Makefile            # 构建自动化
├── *_test.go           # 测试文件
├── Dockerfile.test     # Docker 测试环境
└── README.md           # 您在这里！👋
```

## 🤔 工作原理

1. **会话创建**：将您的描述转换为分支名（例如，"Fix login bug" → `feature/fix-login-bug`）
2. **集中存储**：在 `~/.ccswitch/worktrees/repo-name/session-name` 中创建 worktree - 您的项目保持整洁！
3. **自动导航**：bash 包装器捕获输出并将您 `cd` 到新目录
4. **会话跟踪**：将除主 worktree 外的所有 worktree 列为活跃会话

### 目录结构
```
~/.ccswitch/                      # 所有 ccswitch 数据在您的主目录中
└── worktrees/                    # 集中的 worktree 存储
    ├── my-project/               # 按仓库名称组织
    │   ├── fix-login-bug/        # 各个会话
    │   ├── add-new-feature/
    │   └── refactor-ui/
    └── another-project/
        ├── update-deps/
        └── new-feature/

# 您的项目目录保持完全整洁！
/Users/you/projects/
├── my-project/                   # 只有您的主仓库
└── another-project/              # 没有杂乱！
```

## 🔧 系统要求

- **Go** 1.21 或更高版本（用于构建）
- **Git** 2.20 或更高版本（用于 worktree 支持）
- **Bash** 或 **Zsh**（用于 shell 集成）

## 💡 使用技巧

- 使用描述性的会话名称 - 它们会成为您的分支名！
- 定期清理保持工作区整洁
- 每个 worktree 都是独立的 - 适合测试不同的方法
- 创建新会话时，工具会尊重您当前的分支

## 🐛 故障排除

**"Failed to create worktree"（创建 worktree 失败）**
- 检查分支是否已存在：`git branch -a`
- 确保您在 git 仓库中
- 验证您在父目录中有写入权限

**Shell 集成不工作**
- 确保已导入 bash 包装器
- 检查 `ccswitch` 是否在您的 PATH 中
- 尝试使用完整路径：`/usr/local/bin/ccswitch`

## 📝 许可证

MIT License - 欢迎在您的项目中使用！

## 🤝 贡献

发现了 bug？有想法？欢迎提交 issue 或 PR！

---

用 ❤️ 打造，献给同时处理多个功能的开发者
