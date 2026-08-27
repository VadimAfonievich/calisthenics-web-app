import type {WorkoutExercise} from './api/workouts'

export const russianSeconds=(value:number)=>{
  const mod100=Math.abs(value)%100,mod10=mod100%10
  const word=mod100>=11&&mod100<=14?'секунд':mod10===1?'секунда':mod10>=2&&mod10<=4?'секунды':'секунд'
  return `${value} ${word}`
}

export const exerciseTargetPhrase=(exercise:WorkoutExercise)=>exercise.target_reps!==undefined?`${exercise.target_reps} раз.`:`${russianSeconds(exercise.target_duration_seconds??0)}.`
export const exerciseIntroPhrase=(exercise:WorkoutExercise)=>`${exercise.name}. ${exerciseTargetPhrase(exercise)}`
export const sessionIntroPhrase=(category:string|undefined,exercise:WorkoutExercise)=>`${category==='warmup'?'Начинаем разминку.':'Начинаем тренировку.'} ${exerciseIntroPhrase(exercise)}`
export const nextExercisePhrase=(exercise:WorkoutExercise)=>`Следующее упражнение — ${exercise.name}. ${exerciseTargetPhrase(exercise)}`
export const restPhrase='Отдохните.'
export const preparationPhrase='Приготовьтесь. Пять.'
export const countdownWords:Record<number,string>={1:'Один',2:'Два',3:'Три',4:'Четыре',5:'Пять'}
export const startedPhrase='Начали.'
export const finishedPhrase='Закончили.'
export const completionPhrase=(category:string|undefined,continues:boolean)=>category==='warmup'?(continues?'Разминка завершена. Переходим к основной тренировке.':'Разминка завершена.'):'Тренировка завершена.'
export const voiceTestPhrase='Голосовые подсказки включены.'
