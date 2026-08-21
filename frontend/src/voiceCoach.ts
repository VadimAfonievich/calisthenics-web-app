export const VOICE_COACH_KEY = "calisthenics_voice_coach_enabled";
export const VOICE_COACH_VOICE_KEY = "calisthenics_voice_coach_voice";
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
  voices() { return this.isSupported()?window.speechSynthesis.getVoices().filter(v=>v.lang.toLowerCase().startsWith("ru")):[]; }
  selectedVoice() {
    const voices=this.voices(); let stored=""; try{stored=localStorage.getItem(VOICE_COACH_VOICE_KEY)??""}catch{/* optional */}
    return voices.find(v=>v.voiceURI===stored||v.name===stored)??voices.find(v=>/male|maxim|максим|yuri|юрий|pavel|павел|alexander|александр/i.test(v.name))??voices[0];
  }
  setVoice(identifier:string) { try{localStorage.setItem(VOICE_COACH_VOICE_KEY,identifier)}catch{/* optional */} }
  speak(text:string,key=text,priority:"normal"|"high"="normal") {
    if (!this.enabled || !this.isSupported() || this.lastKey===key) return false;
    if(priority==="high") window.speechSynthesis.cancel();
    this.lastKey=key;
    try {
      const utterance=new SpeechSynthesisUtterance(text); utterance.lang="ru-RU";utterance.rate=1.08;utterance.pitch=.88;
      const russian=this.selectedVoice(); if (russian) utterance.voice=russian;
      window.speechSynthesis.speak(utterance); return true;
    } catch { return false; }
  }
  countdown(scope:string,remaining:number) { if (remaining>=1&&remaining<=5) return this.speak(words[remaining],`${scope}:${remaining}`,"high"); return false; }
  restCountdown(remaining:number) { if (remaining===5) this.speak("Приготовьтесь.","rest-ready"); if (remaining===0) return this.speak("Начали.","rest-start"); return this.countdown("rest",remaining); }
  announceRest(seconds:number) { return this.speak(`Отдых ${seconds} секунд.`,`rest:${seconds}:${Date.now()}`); }
  announceStart() { return this.speak("Начали.",`start:${Date.now()}`,"high"); }
  announceSetComplete() { return this.speak("Подход завершён.",`set-complete:${Date.now()}`,"high"); }
  announceNextExercise(name:string) { return this.speak(`Следующее упражнение: ${name}.`,`next:${name}:${Date.now()}`); }
  announceExercise(name:string,reps?:number) { return this.speak(`${name}.${reps!==undefined?` ${reps} раз.`:""}`,`exercise:${name}:${reps??"time"}`); }
  announceExerciseComplete() { return this.speak("Упражнение завершено.",`exercise-complete:${Date.now()}`); }
  test() { return this.speak("Голосовые подсказки включены.",`test:${Date.now()}`); }
}
