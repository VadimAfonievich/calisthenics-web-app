import {afterEach,describe,expect,it,vi} from "vitest";
import {VoiceCoach} from "./voiceCoach";

class Utterance { lang=""; voice?:SpeechSynthesisVoice; constructor(public text:string){} }
const setup=()=>{const spoken:string[]=[];const speech={speak:vi.fn((u:Utterance)=>spoken.push(u.text)),cancel:vi.fn(),getVoices:vi.fn(()=>[])};vi.stubGlobal("SpeechSynthesisUtterance",Utterance);vi.stubGlobal("window",{speechSynthesis:speech});return {spoken,speech};};
afterEach(()=>vi.unstubAllGlobals());
describe("VoiceCoach",()=>{
  it("falls back safely when speech is unsupported",()=>{vi.stubGlobal("window",{});expect(new VoiceCoach().speak("test")).toBe(false)});
  it("announces only the final five timed seconds",()=>{const {spoken}=setup();const voice=new VoiceCoach();[10,6,5,4,3,2,1].forEach(x=>voice.countdown("set",x));expect(spoken).toEqual(["Пять","Четыре","Три","Два","Один"])});
  it("announces set, rest and next exercise",()=>{const {spoken}=setup();const voice=new VoiceCoach();voice.announceSetComplete();voice.announceRest(60);voice.announceNextExercise("Подтягивания");expect(spoken).toEqual(["Подход завершён.","Отдых 60 секунд.","Следующее упражнение: Подтягивания."])});
  it("announces rest preparation, final countdown and start",()=>{const {spoken}=setup();const voice=new VoiceCoach();[5,4,3,2,1,0].forEach(x=>voice.restCountdown(x));expect(spoken).toEqual(["Приготовьтесь.","Пять","Четыре","Три","Два","Один","Начали."])});
  it("does not duplicate a transition and cancels stale speech",()=>{const {spoken,speech}=setup();const voice=new VoiceCoach();voice.countdown("rest",5);voice.countdown("rest",5);voice.cancel();expect(spoken).toEqual(["Пять"]);expect(speech.cancel).toHaveBeenCalled()});
  it("stays silent while disabled",()=>{const {spoken}=setup();const voice=new VoiceCoach(false);voice.countdown("set",5);expect(spoken).toEqual([])});
});
