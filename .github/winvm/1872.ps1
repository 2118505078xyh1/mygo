# #1872: a page's own full screen takes its window full screen.
. "$env:VM\lib.ps1"
Add-Type @'
using System;
using System.Runtime.InteropServices;
public static class R {
  [StructLayout(LayoutKind.Sequential)] public struct RECT { public int l, t, r, b; }
  [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out RECT r);
  [DllImport("user32.dll")] public static extern int GetWindowLong(IntPtr h, int i);
}
'@
$env:FSCMD = "$env:OUT\fscmd.txt"
$exe = "$env:RUNNER_TEMP\fs.exe"
BuildCheck fullscreen $exe
$scr = [System.Windows.Forms.Screen]::PrimaryScreen.Bounds
"screen $($scr.Width)x$($scr.Height)"
$p = StartApp $exe @() 8
$h = $p.MainWindowHandle
function St($label) {
  $r = New-Object R+RECT; [void][R]::GetWindowRect($h, [ref]$r)
  $style = [R]::GetWindowLong($h, -16)
  $full = ($r.l -le 0 -and $r.t -le 0 -and $r.r -ge $scr.Width -and $r.b -ge $scr.Height)
  "STATE $label rect=$($r.l),$($r.t),$($r.r),$($r.b) caption=$((($style -band 0x00C00000) -eq 0x00C00000)) covers_screen=$full"
}
function Center { $r = New-Object R+RECT; [void][R]::GetWindowRect($h, [ref]$r); FrontWindow $h; Click ([int](($r.l + $r.r) / 2)) ([int](($r.t + $r.b) / 2)) }
function Cmd($c) { Set-Content $env:FSCMD $c; Start-Sleep 2 }
function Esc { FrontWindow $h; Key $VK.ESC; Start-Sleep 3 }

St start; Shot start | Out-Null
# A: a real click on the page asks for full screen.
Center; Start-Sleep 3; St 'A-after-click'; Shot A | Out-Null
# B: Escape leaves the page's full screen; the window follows.
Esc; St 'B-after-esc'; Shot B | Out-Null
# C: re-entry; then the app asks for full screen too, and the page leaving keeps it.
Center; Start-Sleep 3; St 'C-after-click'
Cmd app-full; St 'C-after-app-full'
Esc; St 'C-after-esc (want covers_screen=True)'; Shot C | Out-Null
Cmd app-leave; Start-Sleep 1; St 'C-after-app-leave'
# D: the page is full screen, the app leaves: the page's element leaves too.
Center; Start-Sleep 3; St 'D-after-click'
Cmd app-leave; Start-Sleep 2; St 'D-after-app-leave'; Shot D | Out-Null
Start-Sleep 1
Stop-Process $p -Force -ErrorAction SilentlyContinue
Start-Sleep 1
"---- app stdout"
Get-Content "$env:OUT\fs-stdout.txt"
"---- app stderr"
Get-Content "$env:OUT\fs-stderr.txt" -ErrorAction SilentlyContinue | Select-Object -Last 30
