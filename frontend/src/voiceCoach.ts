import type {WorkoutExercise} from './api/workouts'
import {completionPhrase,countdownWords,finishedPhrase,nextExercisePhrase,preparationPhrase,restPhrase,sessionIntroPhrase,startedPhrase,voiceTestPhrase} from './voiceCoachPhrases'

export const VOICE_COACH_KEY = "calisthenics_voice_coach_enabled";
export const VOICE_COACH_VOICE_KEY = "calisthenics_voice_coach_voice";

export const voiceCoachPreference = () => {
  try { return localStorage.getItem(VOICE_COACH_KEY) !== "false"; } catch { return true; }
};
export const saveVoiceCoachPreference = (enabled:boolean) => {
  try { localStorage.setItem(VOICE_COACH_KEY,String(enabled)); } catch { /* optional preference */ }
};

export class VoiceCoach {
  private lastKey = "";
  private spokenKeys = new Set<string>();
  constructor(private enabled = true) {}
  setEnabled(enabled:boolean) { this.enabled=enabled; if (!enabled) this.cancel(); }
  isSupported() { return typeof window !== "undefined" && "speechSynthesis" in window && typeof SpeechSynthesisUtterance !== "undefined"; }
  cancel() { this.lastKey=""; if (this.isSupported()) window.speechSynthesis.cancel(); }
  dispose() { this.cancel(); this.spokenKeys.clear(); }
  voices() { return this.isSupported()?window.speechSynthesis.getVoices().filter(v=>v.lang.toLowerCase().startsWith("ru")):[]; }
  selectedVoice() {
    const voices=this.voices(); let stored=""; try{stored=localStorage.getItem(VOICE_COACH_VOICE_KEY)??""}catch{/* optional */}
    return voices.find(v=>v.voiceURI===stored||v.name===stored)??voices.find(v=>/male|maxim|максим|yuri|юрий|pavel|павел|alexander|александр/i.test(v.name))??voices[0];
  }
  setVoice(identifier:string) { try{localStorage.setItem(VOICE_COACH_VOICE_KEY,identifier)}catch{/* optional */} }
  speak(text:string,key=text,priority:"normal"|"high"="normal") {
    if (!this.enabled || !this.isSupported() || this.lastKey===key || this.spokenKeys.has(key)) return false;
    if(priority==="high") window.speechSynthesis.cancel();
    this.lastKey=key;this.spokenKeys.add(key);if(this.spokenKeys.size>250)this.spokenKeys.delete(this.spokenKeys.values().next().value!);
    try {
      const utterance=new SpeechSynthesisUtterance(text); utterance.lang="ru-RU";utterance.rate=1.08;utterance.pitch=.88;
      const russian=this.selectedVoice(); if (russian) utterance.voice=russian;
      window.speechSynthesis.speak(utterance); return true;
    } catch { return false; }
  }
  countdown(scope:string,remaining:number) { if (remaining>=1&&remaining<=5) return this.speak(countdownWords[remaining],`${scope}:${remaining}`,"high"); return false; }
  preparationCountdown(scope:string,remaining:number) {
    if (remaining===5) {
      return this.speak(preparationPhrase,`${scope}:ready:5`,"high");
    }
    return this.countdown(scope,remaining);
  }
  announceSessionStart(category:string|undefined,exercise:WorkoutExercise,eventID:string) { return this.speak(sessionIntroPhrase(category,exercise),`session-intro:${eventID}`,"high"); }
  announceTransition(_seconds:number,next:WorkoutExercise|undefined,eventID:string) { return this.speak(next?nextExercisePhrase(next):restPhrase,`transition:${eventID}`,"normal"); }
  announceStart(eventID=`start:${Date.now()}`) { return this.speak(startedPhrase,`start:${eventID}`,"high"); }
  announceFinished(eventID:string) { return this.speak(finishedPhrase,`finished:${eventID}`,"high"); }
  announceCompletion(category:string|undefined,continues:boolean,eventID:string) { return this.speak(completionPhrase(category,continues),`completion:${eventID}`); }
  test() { return this.speak(voiceTestPhrase,`test:${Date.now()}`); }
}
