// @vitest-environment jsdom
import {cleanup,render,screen} from '@testing-library/react'
import {afterEach,describe,expect,it,vi} from 'vitest'
import {WorkoutFields} from './coachAuthoring'
afterEach(cleanup)
describe('standalone Coach workout fields',()=>{it('has warmup control and no program coupling fields',()=>{render(<WorkoutFields value={{title:'Standalone',description:'Plan',difficulty:'beginner',category:'strength',estimated_minutes:20,warmup_enabled:true,exercises:[]}} change={vi.fn()} opts={{categories:[],exercises:[],programs:[],program_levels:[],workouts:[],skills:[],media:[]}}/>);expect(screen.getByLabelText('Разминка перед тренировкой')).toBeTruthy();expect(screen.queryByText('Этап программы')).toBeNull();expect(screen.queryByText('День программы')).toBeNull();expect(screen.queryByText('Выберите программу')).toBeNull()})})
