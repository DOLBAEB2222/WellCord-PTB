package quest

import (
	"fmt"
	"os"
	"path/filepath"
)

const automatorScript = `(function () {
  const delay = (min, max) => new Promise(resolve => {
    const ms = Math.floor(Math.random() * (max - min + 1)) + min;
    setTimeout(resolve, ms);
  });

  class QuestAutomator {
    constructor() {
      this.wpRequire = null;
      this.stores = {};
      this.isRunning = false;
      this.supportedTasks = ["WATCH_VIDEO", "PLAY_ON_DESKTOP", "STREAM_ON_DESKTOP", "PLAY_ACTIVITY", "WATCH_VIDEO_ON_MOBILE"];
    }

    initializeWebpack() {
      delete window.$;
      this.wpRequire = webpackChunkdiscord_app.push([[Symbol()], {}, r => r]);
      webpackChunkdiscord_app.pop();

      this.stores.ApplicationStreamingStore = this.findStore("getStreamerActiveStreamMetadata");
      this.stores.RunningGameStore = this.findStore("getRunningGames");
      this.stores.QuestsStore = this.findStore("getQuest");
      this.stores.ChannelStore = this.findStore("getAllThreadsForParent");
      this.stores.GuildChannelStore = this.findStore("getSFWDefaultChannel");
      this.stores.FluxDispatcher = this.findStore("flushWaitQueue");
      this.stores.api = this.findStore("get");
    }

    findStore(methodName) {
      for (const id in this.wpRequire.c) {
        const mod = this.wpRequire.c[id].exports;
        if (!mod) continue;
        if (mod[methodName]) return mod;
        if (mod.default && mod.default[methodName]) return mod.default;
      }
      return null;
    }

    listQuests() {
      if (!this.stores.QuestsStore) return [];
      const quests = [...this.stores.QuestsStore.quests.values()];
      return quests.filter(x =>
        x.userStatus?.enrolledAt &&
        !x.userStatus?.completedAt &&
        new Date(x.config.expiresAt).getTime() > Date.now() &&
        this.supportedTasks.some(task => Object.keys((x.config.taskConfig ?? x.config.taskConfigV2).tasks).includes(task))
      );
    }

    async executeQuests() {
      if (this.isRunning) return;
      this.isRunning = true;
      try {
        if (!this.wpRequire) this.initializeWebpack();
        const quests = this.listQuests();
        for (const quest of quests) {
          await this.processQuest(quest);
          await delay(1000, 3500);
        }
      } finally {
        this.isRunning = false;
      }
    }

    async processQuest(quest) {
      const taskConfig = quest.config.taskConfig ?? quest.config.taskConfigV2;
      const taskKey = this.supportedTasks.find(task => Object.keys(taskConfig.tasks).includes(task));
      if (!taskKey) return;
      switch (taskKey) {
        case "WATCH_VIDEO":
        case "WATCH_VIDEO_ON_MOBILE":
          await this.simulateWatchVideo(quest, taskKey);
          break;
        case "PLAY_ON_DESKTOP":
          await this.simulatePlayGame(quest);
          break;
        case "STREAM_ON_DESKTOP":
          await this.simulateStream(quest);
          break;
        case "PLAY_ACTIVITY":
          await this.simulateActivity(quest);
          break;
        default:
          break;
      }
    }

    async simulateWatchVideo(quest, taskKey) {
      const task = (quest.config.taskConfig ?? quest.config.taskConfigV2).tasks[taskKey];
      const totalSeconds = task?.durationSeconds ?? 60;
      const step = 10;
      for (let watched = 0; watched < totalSeconds; watched += step) {
        await this.reportProgress(quest.id, taskKey, Math.min(watched + step, totalSeconds));
        await delay(900, 2000);
      }
    }

    async simulatePlayGame(quest) {
      const task = (quest.config.taskConfig ?? quest.config.taskConfigV2).tasks.PLAY_ON_DESKTOP;
      const totalSeconds = task?.durationSeconds ?? 300;
      const step = 30;
      for (let played = 0; played < totalSeconds; played += step) {
        await this.reportProgress(quest.id, "PLAY_ON_DESKTOP", Math.min(played + step, totalSeconds));
        await delay(1000, 2500);
      }
    }

    async simulateStream(quest) {
      const task = (quest.config.taskConfig ?? quest.config.taskConfigV2).tasks.STREAM_ON_DESKTOP;
      const totalSeconds = task?.durationSeconds ?? 300;
      const step = 30;
      for (let streamed = 0; streamed < totalSeconds; streamed += step) {
        await this.reportProgress(quest.id, "STREAM_ON_DESKTOP", Math.min(streamed + step, totalSeconds));
        await delay(1000, 2600);
      }
    }

    async simulateActivity(quest) {
      const task = (quest.config.taskConfig ?? quest.config.taskConfigV2).tasks.PLAY_ACTIVITY;
      const totalSeconds = task?.durationSeconds ?? 180;
      const step = 20;
      for (let played = 0; played < totalSeconds; played += step) {
        await this.reportProgress(quest.id, "PLAY_ACTIVITY", Math.min(played + step, totalSeconds));
        await delay(900, 2200);
      }
    }

    async reportProgress(questId, taskKey, seconds) {
      const endpoint = "/quests/" + questId + "/progress";
      const payload = { taskKey, seconds };
      if (this.stores.api?.post) {
        await this.stores.api.post(endpoint, payload);
      } else if (this.stores.api?.put) {
        await this.stores.api.put(endpoint, payload);
      }
    }
  }

  const automator = new QuestAutomator();
  window.__questAutomator = automator;
  window.__questControl = {
    start: () => automator.executeQuests(),
    list: () => automator.listQuests(),
  };
})();`

func WriteAutomatorScript(destination string) error {
	if destination == "" {
		return fmt.Errorf("missing destination")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	return os.WriteFile(destination, []byte(automatorScript), 0644)
}
