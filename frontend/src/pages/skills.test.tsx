// @vitest-environment jsdom
import {QueryClient,QueryClientProvider} from '@tanstack/react-query'
import {cleanup,render,screen} from '@testing-library/react'
import {MemoryRouter,Route,Routes} from 'react-router-dom'
import type {ReactNode} from 'react'
import {afterEach,beforeEach,describe,expect,it,vi} from 'vitest'
import {useSessionStore} from '../store/session'
import {SkillDetailPage,SkillsPage} from './skills'
const mocks=vi.hoisted(()=>({getSkillMap:vi.fn(),getSkill:vi.fn(),completeSkillLevel:vi.fn(),masterSkill:vi.fn()}))
vi.mock('../api/skills',async original=>({...await original<typeof import('../api/skills')>(),...mocks}))
const id='65000000-0000-0000-0000-000000000001',skill={id,code:'HANDSTAND',name:'Стойка на руках',description:'Прогрессия',category:'SKILL',difficulty:'beginner',icon:'🤸',xp_reward:250,final_criterion_type:'duration_seconds',final_criterion_value:10,status:'in_progress',current_level:2,total_levels:5,progress_percent:20}
const wrap=(ui:ReactNode,path='/skills')=>render(<QueryClientProvider client={new QueryClient({defaultOptions:{queries:{retry:false}}})}><MemoryRouter initialEntries={[path]}>{ui}</MemoryRouter></QueryClientProvider>)
beforeEach(()=>{useSessionStore.setState({accessToken:'token',status:'authenticated'});vi.clearAllMocks()});afterEach(cleanup)
describe('skill UI',()=>{
  it('renders locked and active graph nodes',async()=>{mocks.getSkillMap.mockResolvedValue({nodes:[skill,{...skill,id:'65000000-0000-0000-0000-000000000002',name:'Muscle-Up',status:'locked',progress_percent:0}],requirements:[]});wrap(<SkillsPage/>);expect(await screen.findByText('Стойка на руках')).toBeTruthy();expect(screen.getByText('Закрыт · 0%')).toBeTruthy()})
  it('groups the map and renders actual prerequisite names',async()=>{const base={...skill,id:'65000000-0000-0000-0000-000000000009',name:'База калистеники',map_group:'basic'},handstand={...skill,map_group:'floor'};mocks.getSkillMap.mockResolvedValue({nodes:[base,handstand],requirements:[{skill_id:handstand.id,required_skill_id:base.id,requirement_type:'skill_mastered',requirement_value:0}]});wrap(<SkillsPage/>);expect(await screen.findByText('Базовые навыки')).toBeTruthy();expect(screen.getByText('Пол')).toBeTruthy();expect(screen.getByText('После: База калистеники')).toBeTruthy()})
  it('renders level states and linked workout',async()=>{mocks.getSkill.mockResolvedValue({skill,levels:[{id:'66000000-0000-0000-0000-000000000001',level_number:1,name:'Подготовка',description:'Тело',criterion_type:'workout_completed',criterion_value:1,status:'completed',progress_value:1,workouts:[{id:'54000000-0000-0000-0000-000000000001',title:'Стойка — уровень 1',estimated_minutes:30}]},{id:'66000000-0000-0000-0000-000000000002',level_number:2,name:'Стена',description:'Линия',criterion_type:'duration_seconds',criterion_value:40,status:'available',progress_value:0,workouts:[]}]});wrap(<Routes><Route path="/skills/:id" element={<SkillDetailPage/>}/></Routes>,`/skills/${id}`);expect(await screen.findByText('Стойка — уровень 1 · 30 мин →')).toBeTruthy();expect(screen.getByText('Критерий: 40 сек')).toBeTruthy();expect(screen.getByRole('button',{name:'Подтвердить уровень'})).toBeTruthy()})
})
