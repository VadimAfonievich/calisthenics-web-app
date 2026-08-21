export const VOICE_COACH_KEY = "calisthenics_voice_coach_enabled";
const words: Record<number,string> = {1:"Один",2:"Два",3:"Три",4:"Четыре",5:"Пять"};

export const voiceCoachPreference = () => {
  try { return localStorage.getItem(VOICE_COACH_KEY) !== "false"; } catch { return true; }
};
export const saveVoiceCoachPreference = (enabled:boolean) => {
  try { localStorage.setItem(VOICE_COACH_KEY,String(enabled)); } catch { /* optional preference */ }
};

export class VoiceCoach {
  private lastKey = "";
  constructor(private enabled = true) {}
  setEnabled(enabled:boolean) { this.enabled=enabled; if (!enabled) this.cancel(); }
  isSupported() { return typeof window !== "undefined" && "speechSynthesis" in window && typeof SpeechSynthesisUtterance !== "undefined"; }
  cancel() { this.lastKey=""; if (this.isSupported()) window.speechSynthesis.cancel(); }
  speak(text:string,key=text) {
    if (!this.enabled || !this.isSupported() || this.lastKey===key) return false;
    this.lastKey=key;
    try {
      const utterance=new SpeechSynthesisUtterance(text); utterance.lang="ru-RU";
      const russian=window.speechSynthesis.getVoices().find(v=>v.lang.toLowerCase().startsWith("ru"));
      if (russian) utterance.voice=russian;
      window.speechSynthesis.speak(utterance); return true;
    } catch { return false; }
  }
  countdown(scope:string,remaining:number) { if (remaining>=1&&remaining<=5) return this.speak(words[remaining],`${scope}:${remaining}`); return false; }
  restCountdown(remaining:number) { if (remaining===5) this.speak("Приготовьтесь.","rest-ready"); if (remaining===0) return this.speak("Начали.","rest-start"); return this.countdown("rest",remaining); }
  announceRest(seconds:number) { return this.speak(`Отдых ${seconds} секунд.`,`rest:${seconds}:${Date.now()}`); }
  announceSetComplete() { return this.speak("Подход завершён.",`set-complete:${Date.now()}`); }
  announceNextExercise(name:string) { return this.speak(`Следующее упражнение: ${name}.`,`next:${name}:${Date.now()}`); }
  test() { return this.speak("Голосовые подсказки включены.",`test:${Date.now()}`); }
}
