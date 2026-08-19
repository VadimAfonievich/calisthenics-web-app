// @vitest-environment jsdom
import {afterEach,describe,expect,it,vi} from 'vitest'
import {completeSkillLevel,getSkillMap,masterSkill} from './skills'
const id='65000000-0000-0000-0000-000000000001'
afterEach(()=>vi.unstubAllGlobals())
describe('skills API',()=>{
  it('loads graph nodes and requirements',async()=>{const fetchMock=vi.fn().mockResolvedValue({ok:true,status:200,json:async()=>({nodes:[],requirements:[]})});vi.stubGlobal('fetch',fetchMock);await getSkillMap('token');expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/skills\/map$/),expect.anything())})
  it('sends level criterion value',async()=>{const fetchMock=vi.fn().mockResolvedValue({ok:true,status:200,json:async()=>({status:'completed'})});vi.stubGlobal('fetch',fetchMock);await completeSkillLevel('token',id,2,40);expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/levels\/2\/complete$/),expect.objectContaining({method:'POST',body:'{"value":40}'}))})
  it('sends mastery criterion to an idempotent endpoint',async()=>{const fetchMock=vi.fn().mockResolvedValue({ok:true,status:200,json:async()=>({status:'mastered',xp_earned:250})});vi.stubGlobal('fetch',fetchMock);await masterSkill('token',id,10);expect(fetchMock).toHaveBeenCalledTimes(1)})
})
