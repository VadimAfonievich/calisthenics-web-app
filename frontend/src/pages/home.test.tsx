// @vitest-environment jsdom
import {QueryClient,QueryClientProvider} from '@tanstack/react-query'
import {cleanup,render,screen} from '@testing-library/react'
import {MemoryRouter} from 'react-router-dom'
import {afterEach,beforeEach,describe,expect,it,vi} from 'vitest'
import {useSessionStore} from '../store/session'
import {HomePage} from './home'
const mocks=vi.hoisted(()=>({listPrograms:vi.fn(),listSkills:vi.fn(),getToday:vi.fn()}))
vi.mock('../api/programs',()=>({listPrograms:mocks.listPrograms}))
vi.mock('../api/skills',()=>({listSkills:mocks.listSkills}))
vi.mock('../api/calendar',()=>({getToday:mocks.getToday,occurrenceRoute:vi.fn()}))
beforeEach(()=>{useSessionStore.setState({accessToken:'token',status:'authenticated'});mocks.listSkills.mockResolvedValue({skills:[]});mocks.getToday.mockResolvedValue({occurrences:[]});mocks.listPrograms.mockResolvedValue({programs:[{id:'40000000-0000-0000-0000-000000000001',name:'Базовая сила',workout_count:4}]})})
afterEach(cleanup)
describe('student home programs',()=>{it('places a compact program preview after the workouts entry',async()=>{render(<QueryClientProvider client={new QueryClient()}><MemoryRouter><HomePage/></MemoryRouter></QueryClientProvider>);const training=screen.getByText('ТРЕНИРОВКИ');const programs=await screen.findByText('ПРОГРАММЫ');expect(training.compareDocumentPosition(programs)&Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();expect((await screen.findByText('Базовая сила')).closest('a')?.getAttribute('href')).toBe('/programs/40000000-0000-0000-0000-000000000001');expect(screen.getByText('Все программы →').closest('a')?.getAttribute('href')).toBe('/programs')})})
