import {afterEach,describe,expect,it,vi} from 'vitest'
import {createPlanned,createSchedule,getCalendar,getToday,monthCells,occurrenceRoute,skipPlanned} from './calendar'

const workout='10000000-0000-0000-0000-000000000001',planned='20000000-0000-0000-0000-000000000001'
afterEach(()=>vi.unstubAllGlobals())
const response=(body:unknown={},status=200)=>({ok:status<400,status,json:async()=>body})
describe('calendar API',()=>{
  it('renders a Monday-first month model and valid workout route',()=>{const cells=monthCells(new Date(2026,7,1));expect(cells.slice(0,5)).toEqual([null,null,null,null,null]);expect(occurrenceRoute({date:'2026-08-19',time:'19:00',workout_id:workout,workout_title:'Test',schedule_id:planned,status:'scheduled',category:'SKILL',difficulty:'beginner',estimated_minutes:20})).toContain(`/workouts/${workout}?scheduled_date=2026-08-19`)})
  it('loads range and today occurrences',async()=>{const fetchMock=vi.fn().mockResolvedValue(response({occurrences:[]}));vi.stubGlobal('fetch',fetchMock);await getCalendar('token','2026-08-01','2026-08-31');await getToday('token');expect(fetchMock.mock.calls[0][0]).toContain('/calendar?from=2026-08-01&to=2026-08-31');expect(fetchMock.mock.calls[1][0]).toMatch(/\/calendar\/today$/)})
  it('creates a recurring multi-weekday schedule',async()=>{const fetchMock=vi.fn().mockResolvedValue(response({schedule:{}} ,201));vi.stubGlobal('fetch',fetchMock);await createSchedule('token',{workout_id:workout,weekdays:[1,3,5],preferred_time:'19:00',start_date:'2026-08-19'});expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toMatchObject({weekdays:[1,3,5],preferred_time:'19:00'})})
  it('creates and skips a one-off workout',async()=>{const fetchMock=vi.fn().mockResolvedValueOnce(response({planned_workout:{id:planned}},201)).mockResolvedValueOnce(response({},204));vi.stubGlobal('fetch',fetchMock);await createPlanned('token',{workout_id:workout,scheduled_date:'2026-08-20'});await skipPlanned('token',planned);expect(fetchMock.mock.calls[0][0]).toMatch(/\/planned-workouts$/);expect(fetchMock.mock.calls[1][0]).toMatch(/\/skip$/)})
  it('rejects an undefined workout before navigation',()=>{expect(()=>occurrenceRoute({date:'2026-08-19',workout_id:'undefined',workout_title:'Bad',status:'scheduled',category:'OTHER',difficulty:'beginner',estimated_minutes:1})).toThrow()})
})
