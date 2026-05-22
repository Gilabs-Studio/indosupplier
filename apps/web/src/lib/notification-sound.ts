const DEFAULT_NOTIFICATION_SOUND_PATH = "/sounds/notification.mp3";
const DEFAULT_COOLDOWN_MS = 2000;

let sharedAudio: HTMLAudioElement | null = null;
let sharedAudioSrc = "";
let lastPlayedAt = 0;

interface NotificationSoundOptions {
  src?: string;
  cooldownMs?: number;
}

function getAudio(src: string): HTMLAudioElement {
  if (!sharedAudio || sharedAudioSrc !== src) {
    sharedAudio = new Audio(src);
    sharedAudio.preload = "auto";
    sharedAudioSrc = src;
  }

  return sharedAudio;
}

export function playNotificationSound(options?: NotificationSoundOptions): void {
  if (typeof window === "undefined") {
    return;
  }

  const src = options?.src ?? DEFAULT_NOTIFICATION_SOUND_PATH;
  const cooldownMs = options?.cooldownMs ?? DEFAULT_COOLDOWN_MS;
  const now = Date.now();

  if (now-lastPlayedAt < cooldownMs) {
    return;
  }

  const audio = getAudio(src);
  lastPlayedAt = now;

  try {
    audio.currentTime = 0;
    const playResult = audio.play();
    if (playResult && typeof playResult.catch === "function") {
      playResult.catch(() => {
        // Ignore autoplay rejections; notification UX still works via toast + badge.
      });
    }
  } catch {
    // Ignore playback errors to avoid breaking notification flow.
  }
}
