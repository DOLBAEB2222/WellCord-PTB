# Discord PTB Ultra — Technical Specification (v2.0)

## 1. Базовая архитектура и фундамент

**Основа:** Electron-обёртка над Discord PTB с полной кастомизацией Chromium-инстанса и интегрированным VLESS/Vmess-клиентом.

**Ключевые технические решения:**
- **Изоляция процессов:** Main process (Node.js + Xray-core) + Renderer process (Discord Web) + Plugin Host (iframe sandbox)
- **Управление состоянием:** Redux-подобный store для настроек клиента и VPN-конфигураций
- **Система обновлений:** Автопатчи через GitHub Releases с delta-обновлениями
- **Security-слой:** CSP headers, certificate pinning, VLESS Reality handshake support

---

## 2. Модуль Nitro+ Emulation (ядро)

**Визуальные премиум-фичи:**
- **Avatar Decorations:** кастомные SVG-оверлеи с CSS-анимацией
- **Profile Effects:** particle-системы (Canvas/WebGL) — снег, огонь, кибер-импульсы
- **Animated Avatars:** GIF/APNG/WebP без ограничений
- **Custom Emojis Everywhere:** глобальный эмодзи-пикер с загрузкой своих наборов
- **Server Boost Badge:** визуальный симулятор баджей бустера
- **HD Video Share:** 4K видео через локальное кодирование FFmpeg

---

## 3. VLESS/Vmess VPN Module (Hiddify-Style)

**Поддерживаемые протоколы:**
- **VLESS** + XTLS-Vision / Reality / TCP / WebSocket / gRPC / HTTP/2
- **Vmess** + WebSocket / TCP / HTTP/2 / gRPC
- **Trojan** (опционально)
- **Shadowsocks** (опционально)

**Функционал импорта конфигов:**

1. **Импорт по ссылке:**
   - `vless://`, `vmess://`, `trojan://` ссылки
   - Базовый импорт: `https://example.com/config.txt`
   - Subscription URL (Hiddify-style): автоматическая загрузка списка серверов
   - QR-code сканирование (через нативный модуль)

2. **Форматы импорта:**
   - Одиночные ссылки
   - JSON конфигурации (V2Ray/Xray format)
   - Clash YAML конфиги (автоконвертация)
   - Hiddify Next backup файлы

**UI-компоненты:**
```
┌─────────────────────────────────────────────┐
│ 🔗 Вставьте ссылку или перетащите файл      │
│ [____________________________________] [+]  │
├─────────────────────────────────────────────┤
│ 📋 Subscription URLs:                       │
│ • https://sub.example.com/vless [🔄][🗑️]   │
│ • https://backup.example.com [🔄][🗑️]      │
├─────────────────────────────────────────────┤
│ 🌍 Быстрый выбор сервера:                   │
│ [🇳🇱 NL-1 12ms ▼] [Подключить 🚀]          │
│ ⚡ Auto-select: Включен [✓]                │
└─────────────────────────────────────────────┘
```

**Техническая реализация:**
- **Xray-core integration:** встроенный xray-core бинарник (или sing-box для лёгковесности)
- **Config Manager:** парсинг VLESS URI → Xray JSON config
- **Режимы подключения:**
  - **System Proxy:** глобальный системный прокси (Windows: netsh, macOS: networksetup, Linux: gsettings)
  - **TUN Mode:** виртуальный сетевой адаптер (tun2socks) — весь трафик
  - **SOCKS5/HTTP Proxy:** локальный порт 10808/10809 для ручной настройки
  - **Discord Only:** перехват только Discord процессов (targeted routing)

**Routing rules (Hiddify-style):**
```json
{
  "routing": {
    "rules": [
      {"domain": ["discord.com", "discord.gg"], "outbound": "proxy"},
      {"domain": ["yandex.ru", "vk.com"], "outbound": "direct"},
      {"ip": ["geoip:private"], "outbound": "direct"}
    ]
  }
}
```

**Advanced Features:**
- **Reality handshake:** автогенерация uTLS fingerprint (Chrome, Firefox, Safari, iOS)
- **Load Balancing:** авто-переключение между серверами по latency
- **URL Test:** периодический ping всех серверов для выбора лучшего
- **Statistics:** график трафика (upload/download) в реальном времени

---

## 4. Auto-Quest System (Webpack Injection)

**Архитектура:**
Встроенный инжектор webpack-модулей Discord с автоматическим выполнением квестов.

**Техническая реализация:**
```javascript
// Core Quest Engine (основан на предоставленном коде)
class QuestAutomator {
  constructor() {
    this.wpRequire = null;
    this.stores = {};
    this.isRunning = false;
  }

  initializeWebpack() {
    // Инъекция в webpack runtime
    delete window.$;
    this.wpRequire = webpackChunkdiscord_app.push([[Symbol()], {}, r => r]);
    webpackChunkdiscord_app.pop();

    // Получение необходимых сторов
    this.stores.ApplicationStreamingStore = this.findStore('getStreamerActiveStreamMetadata');
    this.stores.RunningGameStore = this.findStore('getRunningGames');
    this.stores.QuestsStore = this.findStore('getQuest');
    this.stores.ChannelStore = this.findStore('getAllThreadsForParent');
    this.stores.GuildChannelStore = this.findStore('getSFWDefaultChannel');
    this.stores.FluxDispatcher = this.findStore('flushWaitQueue');
    this.stores.api = this.findStore('get');
  }

  async executeQuests() {
    const supportedTasks = [
      "WATCH_VIDEO",
      "PLAY_ON_DESKTOP",
      "STREAM_ON_DESKTOP",
      "PLAY_ACTIVITY",
      "WATCH_VIDEO_ON_MOBILE"
    ];
    const quests = [...this.stores.QuestsStore.quests.values()]
      .filter(x => x.userStatus?.enrolledAt &&
                   !x.userStatus?.completedAt &&
                   new Date(x.config.expiresAt).getTime() > Date.now() &&
                   supportedTasks.some(y =>
                     Object.keys((x.config.taskConfig ?? x.config.taskConfigV2).tasks).includes(y)
                   ));

    for (const quest of quests) {
      await this.processQuest(quest);
    }
  }

  // Методы spoof'инга из предоставленного кода...
}
```

**UI Control Panel:**
```
┌─────────────────────────────────────────────┐
│ 🤖 Auto-Quest Manager                       │
├─────────────────────────────────────────────┤
│ Статус: [🟢 Активен | 🔴 Остановлен]        │
│                                             │
│ Текущие квесты:                             │
│ ☐ Watch Video (2/5 мин) [▶️ Пропустить]    │
│ ☐ Play Game (15/30 мин) [⏳ В процессе]    │
│ ☐ Stream VC (0/10 мин) [👻 Запустить]      │
│                                             │
│ [🚀 Запустить все] [⏹️ Остановить]         │
│ ⚡ Автозапуск при старте: [✓]              │
│ 🎭 Имитировать активность: [✓]             │
└─────────────────────────────────────────────┘
```

**Безопасность:**
- Rate limiting имитации (random delays 1–5 сек)
- Human-like behavior (случайные движения мыши при «просмотре» видео)
- Anti-detection: смена fingerprint, рандомизация PID

---

## 5. UI/UX Customization Engine

**Theme System (6 стилей):**

1. **Cyberpunk 2077** — Neon cyan/magenta, glitch effects, scanlines
2. **Liquid Glass** — iOS 18 style, blur, translucency
3. **Purple Minimalism** — Deep purple, clean whitespace
4. **Vaporwave Sunset** — Retro 80s, grid floors, pink gradients
5. **Tokyo Night** — Dark dev theme, syntax highlighting colors
6. **Nature Zen** — Earth tones, organic shapes

---

## 6. Browser & Яндекс Музыка Integration

- **Встроенный браузер:** Chromium tab внутри клиента
- **Яндекс Музыка Rich Presence:** «Слушает...» с обложкой и кнопкой
- **Overlay Controls:** mini-player поверх Discord
- **Deep Linking:** `discordptb://` protocol handler

---

## 7. Additional Features

- **Message Scheduler:** отложенные сообщения
- **Ghost Mode:** невидимое чтение
- **Multi-Account:** 5+ аккаунтов
- **Backup Manager:** экспорт настроек
- **Keyboard Macros:** Vim-like navigation

---

## 8. Repository Structure

```
discord-ptb-ultra/
├── src/
│   ├── main/
│   │   ├── vpn/           # Xray/sing-box integration
│   │   ├── config/        # VLESS parser & manager
│   │   └── native/        # C++ addons (TUN, etc.)
│   ├── renderer/
│   │   ├── quest-injector/ # Webpack injection engine
│   │   └── themes/
│   └── shared/
├── themes/                # CSS files
├── xray-bin/              # Xray-core binaries (win/mac/linux)
└── build/
```

---

## Ключевые изменения

- ✅ **VLESS/Vmess full support** с импортом по ссылке/subscription
- ✅ **Hiddify-compatible** config management
- ✅ **Auto-quest system** встроен в webpack runtime
- ✅ **Xray-core integration** для максимальной совместимости протоколов
