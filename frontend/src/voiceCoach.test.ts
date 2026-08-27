import {afterEach,describe,expect,it,vi} from 'vitest'
import type {WorkoutExercise} from './api/workouts'
import {VoiceCoach} from './voiceCoach'
import {completionPhrase,russianSeconds} from './voiceCoachPhrases'

class Utterance {lang='';rate=1;pitch=1;voice?:SpeechSynthesisVoice;constructor(public text:string){}}
const setup=()=>{const spoken:string[]=[];const speech={speak:vi.fn((u:Utterance)=>spoken.push(u.text)),cancel:vi.fn(),getVoices:vi.fn(()=>[])};vi.stubGlobal('SpeechSynthesisUtterance',Utterance);vi.stubGlobal('window',{speechSynthesis:speech});return{spoken,speech}}
const timed=(name='Планка',seconds=30):WorkoutExercise=>({id:name,name,sets:2,target_duration_seconds:seconds,rest_seconds:45})
const reps=(name='Отжимания',count=15):WorkoutExercise=>({id:name,name,sets:2,target_reps:count,rest_seconds:60})
const prepare=(voice:VoiceCoach,key:string)=>[5,4,3,2,1].forEach(value=>voice.preparationCountdown(key,value))
afterEach(()=>vi.unstubAllGlobals())

describe('Voice Coach choreography',()=>{
  it('announces warmup first timed exercise before prepare and start',()=>{const{spoken}=setup(),voice=new VoiceCoach();voice.announceSessionStart('warmup',timed('Круговые движения плечами'),'session');prepare(voice,'first');voice.announceStart('first');expect(spoken).toEqual(['Начинаем разминку. Круговые движения плечами. 30 секунд.','Приготовьтесь. Пять.','Четыре','Три','Два','Один','Начали.'])})
  it('announces warmup first reps exercise',()=>{const{spoken}=setup(),voice=new VoiceCoach();voice.announceSessionStart('warmup',reps(),'session');expect(spoken).toEqual(['Начинаем разминку. Отжимания. 15 раз.'])})
  it('keeps same timed exercise as a short rest prompt without duration',()=>{const{spoken}=setup(),voice=new VoiceCoach();voice.announceFinished('set-1');voice.announceTransition(45,undefined,'set-1');prepare(voice,'set-2');voice.announceStart('set-2');expect(spoken).toEqual(['Закончили.','Отдохните.','Приготовьтесь. Пять.','Четыре','Три','Два','Один','Начали.'])})
  it('counts down the timed ending before finished',()=>{const{spoken}=setup(),voice=new VoiceCoach();[5,4,3,2,1].forEach(value=>voice.countdown('ending:set-1',value));voice.announceFinished('set-1');expect(spoken).toEqual(['Пять','Четыре','Три','Два','Один','Закончили.'])})
  it.each([['timed',timed('Круговые движения тазом'), '30 секунд.'],['reps',reps('Приседания'), '15 раз.']])('announces next %s exercise and target without rest duration',(_,next,target)=>{const{spoken}=setup(),voice=new VoiceCoach();voice.announceFinished('transition');voice.announceTransition(45,next as WorkoutExercise,'transition');expect(spoken).toEqual(['Закончили.',`Следующее упражнение — ${(next as WorkoutExercise).name}. ${target}`])})
  it('handles reps to timed and reps to same reps set',()=>{const{spoken}=setup(),voice=new VoiceCoach();voice.announceTransition(60,timed(),'reps-timed');voice.announceTransition(60,undefined,'reps-same');expect(spoken).toEqual(['Следующее упражнение — Планка. 30 секунд.','Отдохните.'])})
  it('does not announce zero-second rest',()=>{const{spoken}=setup(),voice=new VoiceCoach();voice.announceTransition(0,timed(),'zero');expect(spoken).toEqual(['Следующее упражнение — Планка. 30 секунд.'])})
  it('announces final workout and warmup completion variants',()=>{const{spoken}=setup(),voice=new VoiceCoach();voice.announceFinished('main');voice.announceCompletion('strength',false,'main');voice.announceCompletion('warmup',false,'warmup');voice.announceCompletion('warmup',true,'flow');expect(spoken).toEqual(['Закончили.','Тренировка завершена.','Разминка завершена.','Разминка завершена. Переходим к основной тренировке.'])})
  it('deduplicates semantic events even after cancellation',()=>{const{spoken}=setup(),voice=new VoiceCoach();voice.announceFinished('same');voice.cancel();voice.announceFinished('same');expect(spoken).toEqual(['Закончили.'])})
  it('allows StrictMode remount replay after disposal cancelled the first intro',()=>{const{spoken}=setup(),voice=new VoiceCoach();voice.announceSessionStart('warmup',timed(),'session');voice.dispose();voice.announceSessionStart('warmup',timed(),'session');expect(spoken).toHaveLength(2)})
  it('cancels stale normal speech before critical countdown',()=>{const{speech}=setup(),voice=new VoiceCoach();voice.announceTransition(60,undefined,'rest');voice.preparationCountdown('next',5);expect(speech.cancel).toHaveBeenCalled()})
  it('stays silent while disabled or unsupported',()=>{const{spoken}=setup();new VoiceCoach(false).announceSessionStart('warmup',timed(),'off');expect(spoken).toEqual([]);vi.stubGlobal('window',{});expect(new VoiceCoach().announceFinished('unsupported')).toBe(false)})
})

it('formats Russian seconds',()=>{expect([1,2,5,11,21,22,25].map(russianSeconds)).toEqual(['1 секунда','2 секунды','5 секунд','11 секунд','21 секунда','22 секунды','25 секунд']);expect(completionPhrase('warmup',true)).toContain('основной тренировке')})
