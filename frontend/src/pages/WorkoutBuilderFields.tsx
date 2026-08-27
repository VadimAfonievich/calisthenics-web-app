import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { contentDetail, type BuilderExercise, type BuilderInput, type CoachOptions, type Option } from "../api/coach";
import { useSessionStore } from "../store/session";

const difficulty: Record<string, string> = { beginner: "Начальный", intermediate: "Средний", advanced: "Продвинутый" };
const equipment: Record<string, string> = { none: "Без оборудования", floor: "Пол", wall: "Стена", "pull-up-bar": "Турник", "parallel-bars": "Брусья", rings: "Кольца", "resistance-band": "Резинка", bench: "Скамья", box: "Тумба" };
const muscles: Record<string, string> = { chest: "Грудь", back: "Спина", shoulders: "Плечи", biceps: "Бицепс", triceps: "Трицепс", forearms: "Предплечья", core: "Кор", glutes: "Ягодицы", quadriceps: "Квадрицепс", hamstrings: "Задняя поверхность бедра", calves: "Икры", "full-body": "Всё тело" };
const directions: Record<string, string> = { "warm-up": "Разминка", push: "Push", pull: "Pull", core: "Core", squat: "Ноги", handstand: "Handstand", planche: "Planche", "front-lever": "Front Lever", "back-lever": "Back Lever", "muscle-up": "Muscle-Up", "l-sit": "L-Sit" };
const presets = { sets: [2,3,4,5], rest: [30,45,60,90,120], reps: [5,8,10,12,15,20], duration: [20,30,45,60] };

export const estimateWorkoutMinutes = (items: BuilderExercise[]) => Math.max(1, Math.ceil(items.reduce((seconds,item)=>seconds + item.sets*((item.target_duration_seconds ?? ((item.target_reps ?? 10)*3))) + Math.max(0,item.sets-1)*item.rest_seconds,0)/60));
export const normalizeWorkoutExercises = (items: BuilderExercise[]) => items.map((item,sort_order)=>({...item,sort_order}));
export const moveWorkoutExercise = (items: BuilderExercise[],index:number,direction:-1|1) => { const target=index+direction;if(target<0||target>=items.length)return normalizeWorkoutExercises(items);const next=[...items];[next[index],next[target]]=[next[target],next[index]];return normalizeWorkoutExercises(next) };
export const filterExerciseOptions = (items:Option[], filters:{search:string;difficulty:string;type:string;equipment:string;direction:string;muscle:string;ownership:string}) => items.filter(item=>{
  const search=filters.search.trim().toLowerCase();
  return (!search || [item.name,item.description,...(item.tags??[])].join(" ").toLowerCase().includes(search)) &&
    (!filters.difficulty||item.difficulty===filters.difficulty) &&
    (!filters.type||(filters.type==="duration"?item.movement_type==="duration":item.movement_type!=="duration")) &&
    (!filters.equipment||(item.equipment??[]).includes(filters.equipment)) &&
    (!filters.direction||(item.tags??[]).includes(filters.direction)) &&
    (!filters.muscle||(item.muscle_groups??[]).includes(filters.muscle)) &&
    (!filters.ownership||(filters.ownership==="own"?!!item.owner_user_id:!item.owner_user_id));
});
type ExerciseFilters={search:string;difficulty:string;type:string;equipment:string;direction:string;muscle:string;ownership:string};
const emptyFilters:ExerciseFilters={search:"",difficulty:"",type:"",equipment:"",direction:"",muscle:"",ownership:""};

type Props={value:BuilderInput;change:(value:BuilderInput)=>void;opts?:CoachOptions};
export function WorkoutBuilderFields({value,change,opts}:Props){
  const list=value.exercises??[], byID=new Map((opts?.exercises??[]).map(x=>[x.id,x]));
  const [picker,setPicker]=useState(false),[pending,setPending]=useState<string[]>([]),[preview,setPreview]=useState<Option>(),[selectedOnly,setSelectedOnly]=useState(false),[workoutPreview,setWorkoutPreview]=useState(false),[warmupPreview,setWarmupPreview]=useState(false);
  const [filters,setFilters]=useState<ExerciseFilters>(emptyFilters);
  const availableEquipment=useMemo(()=>Array.from(new Set((opts?.exercises??[]).flatMap(x=>x.equipment??[]))).sort(),[opts?.exercises]);
  const availableMuscles=useMemo(()=>Array.from(new Set((opts?.exercises??[]).flatMap(x=>x.muscle_groups??[]))).sort(),[opts?.exercises]);
  const availableDirections=useMemo(()=>Object.keys(directions).filter(tag=>(opts?.exercises??[]).some(x=>x.tags?.includes(tag))),[opts?.exercises]);
  const update=(index:number,patch:Partial<BuilderExercise>)=>change({...value,exercises:list.map((item,i)=>i===index?{...item,...patch}:item)});
  const addPending=()=>{const used=new Set(list.map(x=>x.exercise_id));const additions=pending.filter(id=>!used.has(id)).map((id,index)=>{const option=byID.get(id);const timed=option?.movement_type==="duration";return {exercise_id:id,sets:3,target_reps:timed?undefined:10,target_duration_seconds:timed?30:undefined,rest_seconds:60,sort_order:list.length+index}});change({...value,exercises:[...list,...additions]});setPending([]);setPicker(false)};
  const selectedWarmup=opts?.warmups.find(x=>x.id===value.warmup_workout_id);
  const token=useSessionStore(s=>s.accessToken);
  const warmupDetail=useQuery({queryKey:["coach-warmup-preview",selectedWarmup?.id],queryFn:()=>contentDetail(token!,"workouts",selectedWarmup!.id),enabled:!!token&&!!selectedWarmup&&warmupPreview});
  useEffect(()=>{if(!picker)return;const previous=document.body.style.overflow;document.body.style.overflow="hidden";return()=>{document.body.style.overflow=previous}},[picker]);
  return <section className="stack workout-builder">
    <section className="builder-section"><p className="eyebrow">1 · Основные настройки</p><label>Ориентировочная длительность, минут<input inputMode="numeric" type="number" min="1" value={value.estimated_minutes??""} onChange={event=>change({...value,estimated_minutes:event.target.value===""?undefined:Number(event.target.value)})}/></label><small>Можно использовать автоматический расчёт после добавления упражнений или скорректировать вручную.</small></section>
    <section className="builder-section">
      <div className="section-heading"><div><p className="eyebrow">2–4</p><h3>Упражнения</h3></div><button type="button" className="secondary-button" onClick={()=>setPicker(true)}>+ Добавить упражнение</button></div>
      {!list.length&&<div className="empty-state compact"><b>Упражнения ещё не добавлены</b><p>Выберите несколько упражнений из библиотеки.</p></div>}
      {list.map((item,index)=><ExerciseConfig key={item.exercise_id} item={item} option={byID.get(item.exercise_id)} index={index} total={list.length} update={patch=>update(index,patch)} move={direction=>change({...value,exercises:moveWorkoutExercise(list,index,direction)})} remove={()=>change({...value,exercises:normalizeWorkoutExercises(list.filter((_,i)=>i!==index))})}/>) }
      {!!list.length&&<div className="duration-estimate"><span>Расчёт по составу: ~{estimateWorkoutMinutes(list)} мин</span><button type="button" onClick={()=>change({...value,estimated_minutes:estimateWorkoutMinutes(list)})}>Использовать</button></div>}
    </section>
    <section className="builder-section"><p className="eyebrow">5</p><h3>Разминка</h3>
      {value.category!=="warmup"&&<label className="toggle-field"><span><b>Разминка перед тренировкой</b><small>Можно выбрать опубликованную разминку или использовать стандартную.</small></span><input aria-label="Разминка перед тренировкой" type="checkbox" checked={value.warmup_enabled??true} onChange={e=>change({...value,warmup_enabled:e.target.checked,warmup_workout_id:e.target.checked?value.warmup_workout_id:undefined})}/></label>}
      {value.category!=="warmup"&&(value.warmup_enabled??true)&&<div className="warmup-selector"><label>Выберите разминку<select value={value.warmup_workout_id??""} onChange={e=>change({...value,warmup_workout_id:e.target.value||undefined})}><option value="">Стандартная разминка</option>{opts?.warmups.map(x=><option key={x.id} value={x.id}>{x.name} · {x.minutes||"—"} мин · {x.exercise_count||0} упр.</option>)}</select></label>{selectedWarmup&&<button type="button" onClick={()=>setWarmupPreview(!warmupPreview)}>Посмотреть разминку</button>}</div>}
      {warmupPreview&&selectedWarmup&&<div className="card inline-preview"><b>{selectedWarmup.name}</b><p>{selectedWarmup.minutes} мин · {selectedWarmup.exercise_count} упражнений</p>{warmupDetail.isLoading&&<p>Загружаем состав…</p>}{Array.isArray(warmupDetail.data?.item.exercises)&&<ol>{(warmupDetail.data!.item.exercises as Array<Record<string,unknown>>).map((x,i)=><li key={i}>{byID.get(String(x.exercise_id))?.name??"Упражнение"}</li>)}</ol>}</div>}
    </section>
    <section className="builder-section"><p className="eyebrow">6</p><h3>Предпросмотр</h3><button type="button" className="secondary-button" onClick={()=>setWorkoutPreview(true)}>Предпросмотр тренировки</button></section>
    {picker&&<ExercisePicker
      items={opts?.exercises??[]} pending={pending} existing={new Set(list.map(x=>x.exercise_id))}
      filters={filters} setFilters={setFilters} equipmentValues={availableEquipment}
      muscleValues={availableMuscles} directionValues={availableDirections}
      selectedOnly={selectedOnly} setSelectedOnly={setSelectedOnly}
      toggle={id=>setPending(current=>current.includes(id)?current.filter(x=>x!==id):[...current,id])}
      preview={setPreview} close={()=>{setPicker(false);setPending([])}} confirm={addPending}
    />}
    {preview&&<ExerciseDetail option={preview} close={()=>setPreview(undefined)} add={()=>{if(!list.some(x=>x.exercise_id===preview.id)&&!pending.includes(preview.id))setPending([...pending,preview.id]);setPreview(undefined)}}/>}
    {workoutPreview&&<WorkoutPreview
      value={value} items={list.map(x=>({item:x,option:byID.get(x.exercise_id)}))}
      warmup={selectedWarmup} close={()=>setWorkoutPreview(false)}
    />}
  </section>
}

function ExercisePicker({items,pending,existing,filters,setFilters,equipmentValues,muscleValues,directionValues,selectedOnly,setSelectedOnly,toggle,preview,close,confirm}:{items:Option[];pending:string[];existing:Set<string>;filters:ExerciseFilters;setFilters:(x:ExerciseFilters)=>void;equipmentValues:string[];muscleValues:string[];directionValues:string[];selectedOnly:boolean;setSelectedOnly:(x:boolean)=>void;toggle:(id:string)=>void;preview:(x:Option)=>void;close:()=>void;confirm:()=>void}){
  const [filterSheet,setFilterSheet]=useState(false);
  const set=(name:keyof ExerciseFilters,value:string)=>setFilters({...filters,[name]:value});
  const reset=()=>setFilters(emptyFilters);
  const activeCount=[filters.difficulty,filters.type,filters.equipment,filters.direction,filters.muscle].filter(Boolean).length;
  const matched=filterExerciseOptions(items,filters).filter(x=>!selectedOnly||pending.includes(x.id));
  const quick=["","warm-up","push","pull","core","squat"].filter(value=>!value||directionValues.includes(value));
  return <div className="builder-overlay picker-overlay" role="dialog" aria-label="Выбор упражнений">
    <div className="builder-sheet picker-sheet">
      <div className="picker-header">
        <header><button type="button" onClick={close} aria-label="Назад">←</button><h2>Выбрать упражнения</h2></header>
        <input autoFocus type="search" placeholder="Поиск упражнения..." value={filters.search} onChange={e=>set("search",e.target.value)}/>
        <div className="picker-main-actions">
          <button type="button" className="filter-action" onClick={()=>setFilterSheet(true)}>Фильтры{activeCount?` · ${activeCount}`:""}</button>
          <div className="ownership-toggle" aria-label="Принадлежность">
            {[["","Все"],["system","Стандартные"],["own","Мои"]].map(([value,label])=><button type="button" key={value} className={filters.ownership===value?"active":""} onClick={()=>set("ownership",value)}>{label}</button>)}
          </div>
        </div>
        <div className="quick-filter-row" aria-label="Быстрые фильтры">
          {quick.map(value=><button type="button" key={value||"all"} className={filters.direction===value?"active":""} onClick={()=>set("direction",value)}>{value?directions[value]:"Все"}</button>)}
        </div>
      </div>
      <div className="picker-results" data-scroll-container="true">
        {matched.map(x=>{const selected=pending.includes(x.id),alreadyUsed=existing.has(x.id);return <article className={`exercise-picker-card ${selected||alreadyUsed?"selected":""}`} key={x.id} onClick={()=>preview(x)}>
          <div className="exercise-card-content"><div><span className="ownership-badge">{x.owner_user_id?"Моё":"Стандартное"}</span>{x.has_demo&&<span className="demo-badge">▶ Demo</span>}</div><h3>{x.name}</h3><p>{difficulty[x.difficulty??""]??x.difficulty} · {(x.equipment??[]).slice(0,1).map(v=>equipment[v]??v).join(", ")||"Без оборудования"}</p><small>{(x.muscle_groups??[]).slice(0,2).map(v=>muscles[v]??v).join(" · ")}</small><div className="tag-row">{(x.tags??[]).slice(0,2).map(tag=><span key={tag}>{directions[tag]??tag}</span>)}</div></div>
          <button type="button" disabled={alreadyUsed} onClick={event=>{event.stopPropagation();toggle(x.id)}}>{alreadyUsed||selected?"✓ Добавлено":"+ Добавить"}</button>
        </article>})}
        {!matched.length&&<div className="empty-state picker-empty"><b>Ничего не найдено</b><p>Измените запрос или сбросьте фильтры.</p><button type="button" onClick={reset}>Сбросить фильтры</button><Link to="/coach/exercises/new/edit">+ Создать своё упражнение</Link></div>}
      </div>
      <footer className="picker-footer"><div><b>Выбрано: {pending.length}</b>{!!pending.length&&<button type="button" onClick={()=>setSelectedOnly(!selectedOnly)}>{selectedOnly?"Показать все":"Показать выбранные"}</button>}</div><button type="button" className="primary-button" disabled={!pending.length} onClick={confirm}>Добавить {pending.length} упражнений</button></footer>
    </div>
    {filterSheet&&<ExerciseFilterSheet items={items} filters={filters} equipmentValues={equipmentValues} muscleValues={muscleValues} directionValues={directionValues} close={()=>setFilterSheet(false)} apply={next=>{setFilters(next);setFilterSheet(false)}}/>}
  </div>
}

function ExerciseFilterSheet({items,filters,equipmentValues,muscleValues,directionValues,close,apply}:{items:Option[];filters:ExerciseFilters;equipmentValues:string[];muscleValues:string[];directionValues:string[];close:()=>void;apply:(filters:ExerciseFilters)=>void}){
  const [draft,setDraft]=useState(filters),set=(name:keyof ExerciseFilters,value:string)=>setDraft({...draft,[name]:value});
  const count=filterExerciseOptions(items,draft).length;
  const clear=()=>setDraft({...emptyFilters,search:filters.search,ownership:filters.ownership});
  return <div className="filter-sheet-overlay" role="dialog" aria-label="Фильтры упражнений"><div className="filter-sheet"><header><h2>Фильтры</h2><button type="button" onClick={close} aria-label="Закрыть фильтры">×</button></header><div className="filter-sheet-body">
    <label>Сложность<select aria-label="Сложность" value={draft.difficulty} onChange={e=>set("difficulty",e.target.value)}><option value="">Любая</option>{Object.entries(difficulty).map(([v,l])=><option value={v} key={v}>{l}</option>)}</select></label>
    <label>Тип<select aria-label="Тип" value={draft.type} onChange={e=>set("type",e.target.value)}><option value="">Любой</option><option value="reps">Повторения</option><option value="duration">Время</option></select></label>
    <label>Оборудование<select aria-label="Оборудование" value={draft.equipment} onChange={e=>set("equipment",e.target.value)}><option value="">Любое</option>{equipmentValues.map(v=><option value={v} key={v}>{equipment[v]??v}</option>)}</select></label>
    <label>Направление<select aria-label="Направление" value={draft.direction} onChange={e=>set("direction",e.target.value)}><option value="">Любое</option>{directionValues.map(v=><option value={v} key={v}>{directions[v]??v}</option>)}</select></label>
    <label>Группа мышц<select aria-label="Группа мышц" value={draft.muscle} onChange={e=>set("muscle",e.target.value)}><option value="">Любая</option>{muscleValues.map(v=><option value={v} key={v}>{muscles[v]??v}</option>)}</select></label>
  </div><footer><button type="button" onClick={clear}>Сбросить</button><button type="button" className="primary-button" onClick={()=>apply(draft)}>Показать {count}</button></footer></div></div>
}
function ExerciseDetail({option,close,add}:{option:Option;close:()=>void;add:()=>void}){return <div className="builder-overlay" role="dialog" aria-label="Описание упражнения"><div className="builder-sheet detail-sheet"><header><h2>{option.name}</h2><button type="button" onClick={close}>×</button></header>{option.has_demo&&<p className="notice">▶ Для упражнения доступна демонстрация</p>}<p>{option.description||"Описание пока не добавлено."}</p>{option.instructions&&<p><b>Техника:</b> {option.instructions}</p>}{option.common_mistakes&&<p><b>Частые ошибки:</b> {option.common_mistakes}</p>}{option.coach_tips&&<p><b>Советы:</b> {option.coach_tips}</p>}<p><b>Мышцы:</b> {(option.muscle_groups??[]).map(v=>muscles[v]??v).join(", ")}</p><p><b>Оборудование:</b> {(option.equipment??[]).map(v=>equipment[v]??v).join(", ")||"не требуется"}</p><button type="button" className="primary-button" onClick={add}>Добавить в тренировку</button></div></div>}
function ExerciseConfig({item,option,index,total,update,move,remove}:{item:BuilderExercise;option?:Option;index:number;total:number;update:(x:Partial<BuilderExercise>)=>void;move:(x:-1|1)=>void;remove:()=>void}){const timed=item.target_duration_seconds!==undefined;const numeric=(raw:string)=>raw===""?undefined:Number(raw);return <article className="workout-exercise-card"><header><div><span className="drag-handle">⠿</span><b>{index+1}. {option?.name??"Упражнение недоступно"}</b><small>{option?.owner_user_id?"Моё":"Стандартное"} · {difficulty[option?.difficulty??""]??option?.difficulty}</small></div><button type="button" className="remove-composition" onClick={()=>confirm("Удалить упражнение из тренировки?")&&remove()}>Удалить</button></header><div className="mode-switch"><button type="button" className={!timed?"active":""} onClick={()=>update({target_reps:10,target_duration_seconds:undefined})}>Повторения</button><button type="button" className={timed?"active":""} onClick={()=>update({target_reps:undefined,target_duration_seconds:30})}>Время</button></div><NumericField label="Подходы" value={item.sets} values={presets.sets} min={1} onChange={value=>update({sets:value as number})}/><NumericField label={timed?"Время, секунд":"Повторения"} value={timed?item.target_duration_seconds:item.target_reps} values={timed?presets.duration:presets.reps} min={1} onChange={value=>update(timed?{target_duration_seconds:value}:{target_reps:value})}/><NumericField label="Отдых после подхода, секунд" value={item.rest_seconds} values={presets.rest} min={0} onChange={value=>update({rest_seconds:value as number})}/><label>Подсказка ученику<textarea value={item.notes??""} onChange={e=>update({notes:e.target.value||undefined})} placeholder="Например: локти держите ближе к корпусу."/></label><div className="row-order-actions"><button type="button" disabled={!index} onClick={()=>move(-1)}>↑ Выше</button><button type="button" disabled={index===total-1} onClick={()=>move(1)}>↓ Ниже</button></div></article>}
function NumericField({label,value,values,min,onChange}:{label:string;value?:number;values:number[];min:number;onChange:(x:number|undefined)=>void}){return <div className="numeric-field"><label>{label}<input inputMode="numeric" type="number" min={min} value={value??""} onChange={e=>onChange(e.target.value===""?undefined:Number(e.target.value))}/></label><div className="preset-row">{values.map(v=><button type="button" key={v} className={value===v?"active":""} onClick={()=>onChange(v)}>{v}</button>)}</div></div>}
function WorkoutPreview({value,items,warmup,close}:{value:BuilderInput;items:Array<{item:BuilderExercise;option?:Option}>;warmup?:Option;close:()=>void}){return <div className="builder-overlay" role="dialog" aria-label="Предпросмотр тренировки"><div className="builder-sheet detail-sheet"><header><div><p className="eyebrow">Предпросмотр ученика</p><h2>{value.title||"Без названия"}</h2></div><button type="button" onClick={close}>×</button></header><p>{value.description||"Описание не заполнено."}</p><p>{difficulty[value.difficulty]} · ~{value.estimated_minutes||estimateWorkoutMinutes(items.map(x=>x.item))} мин · {items.length} упр.</p>{value.warmup_enabled!==false&&<p><b>Разминка:</b> {warmup?.name||"Стандартная"}</p>}<ol className="preview-outline">{items.map(({item,option})=><li key={item.exercise_id}><b>{option?.name}</b><span>{item.sets}×{item.target_duration_seconds?`${item.target_duration_seconds} сек`:item.target_reps} · отдых {item.rest_seconds} сек</span></li>)}</ol></div></div>}
