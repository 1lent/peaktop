# peaktop

> Apple Silicon system monitor for the terminal — real-time CPU, GPU, memory, network, battery, and thermal metrics. Built in Go. No root required for core metrics.
> <img width="1260" height="595" alt="image" src="https://github.com/user-attachments/assets/32063b20-2db1-4d9d-ac21-0264a593b52d" />




## Quick Start

```bash
go install github.com/1lent/peaktop@latest
peaktop
```

Or clone and build:

```bash
git clone https://github.com/1lent/peaktop.git
cd peaktop && make build && ./peaktop
```

## Usage

```bash
peaktop                 # 500ms update interval (default)
peaktop -i 200          # 200ms interval (faster)
peaktop -i 1000         # 1 second interval (slower)
sudo peaktop            # required for temperature, fan, and power metrics
```

## Keybinds

| Key | Action |
|---|---|
| `1`-`5` | Switch tabs |
| `t` | Cycle theme (Dark → Light → Dracula) |
| `+` / `-` | Increase / decrease tick rate |
| `j` / `k` or arrows | Scroll process list |
| `h` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit (with save/discard log prompt) |

## Tabs

**Overview** — CPU/GPU/ANE gauges with sparklines, per-core heatmap blocks, CPU/GPU frequency, VRAM usage, memory pressure bar, network per-interface throughput, thermal state, battery status, FPS counter.

**Processes** — Top 50 processes by CPU% with PID, name, CPU%, and MEM%.

**Thermal** — Thermal pressure level, CPU/GPU temperatures (sudo), fan speeds with activity bars and sparkline (sudo, M-series only).

**Network** — Per-interface RX/TX throughput updated in real-time.

**Battery** — Charge percentage bar, health, cycle count, time remaining, design capacity. Hidden on desktop systems without batteries.

## Comparison

| Feature | peaktop | asitop | btop | htop |
|---|---|---|---|---|
| CPU P/E-core breakdown | ✅ | ✅ | ❌ | ❌ |
| Per-core heatmap blocks | ✅ | ❌ | ❌ | ❌ |
| GPU usage + frequency | ✅ | ✅ | ✅ Basic | ❌ |
| GPU VRAM | ✅ | ❌ | ❌ | ❌ |
| ANE usage | ✅ Chip-dependent | ✅ Via power | ❌ | ❌ |
| Memory + swap | ✅ | ✅ | ✅ | ❌ Swap only |
| Network per-interface | ✅ | ❌ | ✅ | ❌ |
| Disk IO | ✅ | ❌ | ✅ | ❌ |
| Power breakdown (W) | ✅ Sudo | ✅ Sudo | ❌ | ❌ |
| Thermal pressure | ✅ | ❌ | ❌ | ❌ |
| Fan RPM + bars | ✅ Sudo, M-series | ❌ | ❌ | ❌ |
| Battery health + cycles | ✅ | ❌ | ✅ Basic bar | ❌ |
| Process list | ✅ CPU% + MEM% | ❌ | ✅ Full | ✅ Full |
| Alert engine | ✅ | ❌ | ❌ | ❌ |
| CSV session export | ✅ | ❌ | ❌ | ❌ |
| Tabbed interface | ✅ 5 tabs | ❌ Single | ❌ Layout | ❌ |
| Themes | ✅ 3 | ✅ 8 | ✅ 20+ | ❌ |
| Mouse support | ❌ | ❌ | ✅ | ❌ |
| Cross-platform | ❌ Apple Silicon | ❌ macOS only | ✅ All | ✅ All |
| No root for core metrics | ✅ | ❌ Sudo always | ✅ | ✅ |

## Features

- **CPU** — P/E cluster breakdown with core counts (2P+4E=6), per-core percentages with heatmap blocks, frequency, 60-sample sparkline
- **GPU** — Usage gauge, active frequency, temperature, VRAM used/total, sparkline
- **ANE** — Apple Neural Engine utilization (chip-dependent, shows "unavailable" when not exposed)
- **Memory** — Wired/active/compressed bar with breakdown, swap usage, pressure percentage
- **Storage** — Total/used disk capacity (via statfs)
- **Network** — Total throughput plus per-interface RX/TX rates
- **Thermal** — Pressure level with percentage, CPU/GPU temperatures, fan RPM bars with history sparkline
- **Battery** — Charge level, charging status, health (with effective mAh when degraded), cycle count, time remaining
- **Power** — Package/CPU/GPU/ANE/DRAM wattage breakdown (sudo required)
- **Alerts** — Configurable thresholds for thermal, battery, memory, GPU with cooldown periods
- **Themes** — Dark, Light, and Dracula — toggle at runtime or set in config
- **CSV Logging** — All metrics saved to `~/.peaktop/logs/YYYY-MM-DD.csv`, prompt to save or discard on quit

## CSV Logs

Every session writes to `~/.peaktop/logs/`. Each row contains:

`timestamp, cpu%, gpu%, ane%, power_w, cpu_w, gpu_w, ane_w, mem_used, mem_pressure%, swap_used, thermal_pressure, cpu_temp_c, gpu_temp_c, fan_rpm, battery%, charging, network_rx_bps, network_tx_bps`

On quit, press `y` to save or `n` to discard.

## Config

Create `~/.peaktop/config.json`:

```json
{"theme": "dracula"}
```

Valid themes: `dark`, `light`, `dracula`. You can also cycle themes with the `t` key at runtime.

## Requirements

| Requirement | Details |
|---|---|
| OS | macOS 13+ |
| Architecture | Apple Silicon (arm64) |
| Go | 1.21+ (build from source only) |
| `sudo` | Only needed for temperature, fan, and power metrics |

## Platform Notes

**M-series Macs** (M1/M2/M3/M4) — Full feature set. Run with `sudo` for temperatures, fan speeds, and power metrics.

**A-series Neo devices** (A14-A18 Pro) — CPU, GPU, memory, network, disk, battery, and alerts work without sudo. Temperature sensors, thermal pressure (`kern.thermalpressure`), and fan data are unavailable on these chips (no fans, sensors not exposed via powermetrics). ANE may show "unavailable" if performance counters aren't exposed. Process visibility may be limited. *(The screenshot above was taken on a Neo — temps and fans are absent by design, not a bug.)*

## License

MIT

## Acknowledgements

Inspired by and built with reference to these excellent tools:

- [asitop](https://github.com/tlkh/asitop) — The original Apple Silicon CLI monitor (Python)
- [btop](https://github.com/aristocratos/btop) — Beautiful cross-platform resource monitor (C++)
- [htop](https://github.com/htop-dev/htop) — The classic interactive process viewer (C)
- [mactop](https://github.com/context-labs/mactop) — Apple Silicon monitor with process-level GPU metrics (Go)
- [nvtop](https://github.com/Syllo/nvtop) — GPU process monitor for NVIDIA/AMD/Intel
- [powermetrics](https://www.unix.com/man-page/osx/1/powermetrics/) — Apple's built-in hardware performance counter utility
