# VHS tape for the cast demo GIF. Render with: sh docs/demo/render.sh
# (render.sh builds cast and puts it first on PATH. `cast build` is deterministic
# and agent-free, so the demo never calls a real agent.)
Output docs/demo.gif

Require cast

Set Shell "bash"
Set FontSize 20
Set Width 1200
Set Height 760
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms

Sleep 600ms
Type "cast targets"
Sleep 500ms
Enter
Sleep 1.5s

Type "cast build examples/changelog --target codex,cursor"
Sleep 700ms
Enter
Sleep 2.5s

Type "find dist -type f | sort"
Sleep 500ms
Enter
Sleep 3s
