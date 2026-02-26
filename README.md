# ltool
lustre GUI tool with fyne/v2


CompileFlags:
  Add: [
    "-Wno-unknown-warning-option",
    "-Wno-unused-command-line-argument",
    # 核心：消除位域截断告警
    "-Wno-bitfield-constant-conversion",
    "-Wno-implicit-int-conversion"
  ]
  # 告诉 clangd 移除不兼容的 GCC 参数
  Remove: [
    "-mpreferred-stack-boundary=*",
    "-mindirect-branch=*",
    "-mindirect-branch-register",
    "-mno-fp-ret-in-387",
    "-fno-allow-store-data-races",
    "-fconserve-stack",
    "-mskip-rax-setup",
    "-mrecord-mcount"
  ]
