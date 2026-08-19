// @vitest-environment jsdom
import {QueryClient,QueryClientProvider} from '@tanstack/react-query'
import {cleanup,render,screen,waitFor} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type {ReactNode} from 'react'
import {MemoryRouter,Route,Routes} from 'react-router-dom'
import {afterEach,beforeEach,describe,expect,it,vi} from 'vitest'
import {useSessionStore} from '../store/session'
import {WorkoutCatalogPage,WorkoutPreviewPage} from './workout'

const mocks=vi.hoisted(()=>({listWorkouts:vi.fn(),getWorkout:vi.fn(),start:vi.fn()}))
vi.mock('../api/workouts',async importOriginal=>({...await importOriginal<typeof import('../api/workouts')>(),...mocks,getSession:vi.fn(),saveSet:vi.fn(),complete:vi.fn()}))
const workout={id:'50000000-0000-0000-0000-000000000001',title:'Стойка на руках — уровень 1',description:'Подготовительная тренировка',estimated_minutes:30,difficulty:'beginner',program_id:'40000000-0000-0000-0000-000000000001',program_name:'Путь к стойке на руках',exercises:[{id:'30000000-0000-0000-0000-000000000001',name:'Hollow Body Hold',sets:3,target_duration_seconds:20,rest_seconds:45}]}
const wrapper=(ui:ReactNode,path='/workouts')=>render(<QueryClientProvider client={new QueryClient({defaultOptions:{queries:{retry:false},mutations:{retry:false}}})}><MemoryRouter initialEntries={[path]}>{ui}</MemoryRouter></QueryClientProvider>)

beforeEach(()=>{useSessionStore.setState({accessToken:'token',status:'authenticated'});vi.clearAllMocks()})
afterEach(cleanup)

describe('workout pages',()=>{
  it('renders the workout catalog without starting a session',async()=>{mocks.listWorkouts.mockResolvedValue({workouts:[{...workout,exercise_count:8,status:'started'}]});wrapper(<WorkoutCatalogPage/>);expect(await screen.findByText(workout.title)).toBeTruthy();expect(screen.getByText('8 упр.')).toBeTruthy();expect(mocks.start).not.toHaveBeenCalled()})
  it('renders preview and starts only after the button click',async()=>{mocks.getWorkout.mockResolvedValue({workout});mocks.start.mockResolvedValue({session:{id:'70000000-0000-0000-0000-000000000001'}});wrapper(<Routes><Route path="/workouts/:id" element={<WorkoutPreviewPage/>}/><Route path="/workout-session/:id" element={<p>Active player</p>}/></Routes>,`/workouts/${workout.id}`);expect(await screen.findByText('Hollow Body Hold')).toBeTruthy();expect(mocks.start).not.toHaveBeenCalled();await userEvent.click(screen.getByRole('button',{name:'Начать тренировку'}));await waitFor(()=>expect(mocks.start).toHaveBeenCalledTimes(1));expect(await screen.findByText('Active player')).toBeTruthy()})
  it('shows an API failure and prevents duplicate clicks while pending',async()=>{mocks.getWorkout.mockResolvedValue({workout});let reject!:(error:Error)=>void;mocks.start.mockReturnValue(new Promise((_,r)=>{reject=r}));wrapper(<Routes><Route path="/workouts/:id" element={<WorkoutPreviewPage/>}/></Routes>,`/workouts/${workout.id}`);const button=await screen.findByRole('button',{name:'Начать тренировку'}) as HTMLButtonElement;await userEvent.click(button);expect(button.disabled).toBe(true);await userEvent.click(button);expect(mocks.start).toHaveBeenCalledTimes(1);reject(new Error('API unavailable'));expect(await screen.findByText('API unavailable')).toBeTruthy()})
})
