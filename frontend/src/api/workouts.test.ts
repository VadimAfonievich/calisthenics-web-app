// @vitest-environment jsdom
import {afterEach,describe,expect,it,vi} from 'vitest'
import {getSession,listWorkouts,saveSet} from './workouts'

const id='50000000-0000-0000-0000-000000000001'
afterEach(()=>vi.unstubAllGlobals())

describe('workout API',()=>{
  it('loads the catalog',async()=>{const fetchMock=vi.fn().mockResolvedValue({ok:true,status:200,json:async()=>({workouts:[]})});vi.stubGlobal('fetch',fetchMock);await listWorkouts('token');expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/workouts$/),expect.anything())})
  it('loads owned session state for resume',async()=>{const fetchMock=vi.fn().mockResolvedValue({ok:true,status:200,json:async()=>({session:{},workout:{},completed_sets:[]})});vi.stubGlobal('fetch',fetchMock);await getSession('token',id);expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(new RegExp(`/workout-sessions/${id}$`)),expect.anything())})
  it('accepts 204 when a set is saved',async()=>{const fetchMock=vi.fn().mockResolvedValue({ok:true,status:204});vi.stubGlobal('fetch',fetchMock);await expect(saveSet('token',id,{exercise_id:id,set_number:1,reps:5,completed:true})).resolves.toBeUndefined();expect(fetchMock).toHaveBeenCalledTimes(1)})
  it('surfaces an API failure',async()=>{vi.stubGlobal('fetch',vi.fn().mockResolvedValue({ok:false,status:500,json:async()=>({error:{code:'FAILED',message:'Set failed'}})}));await expect(saveSet('token',id,{exercise_id:id,set_number:1,reps:5,completed:true})).rejects.toMatchObject({code:'FAILED',status:500})})
})
