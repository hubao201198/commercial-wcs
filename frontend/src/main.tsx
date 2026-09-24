import React,{useEffect,useMemo,useState} from 'react';
import {createRoot} from 'react-dom/client';
import './style.css';

type Row=Record<string,any>;
const tabs=['总览','入库 ASN','库存','出库订单','作业任务','AI Agent','审计日志'];

async function api(path:string,opt:RequestInit={}){
  const r=await fetch(path,{...opt,headers:{'Content-Type':'application/json','X-Actor':'demo.supervisor',...(opt.headers||{})}});
  const data=await r.json().catch(()=>({}));
  if(!r.ok) throw new Error(data.error||('HTTP '+r.status));
  return data;
}

function App(){
  const [tab,setTab]=useState('总览');
  const [dash,setDash]=useState<Row>({});
  const [rows,setRows]=useState<Row[]>([]);
  const [message,setMessage]=useState('');
  const [busy,setBusy]=useState(false);
  const [aiPlan,setAiPlan]=useState<any>(null);

  const load=async()=>{
    try{
      setDash(await api('/api/dashboard'));
      if(tab==='入库 ASN') setRows(await api('/api/asns'));
      else if(tab==='库存') setRows(await api('/api/inventory'));
      else if(tab==='出库订单') setRows(await api('/api/orders'));
      else if(tab==='作业任务') setRows(await api('/api/tasks'));
      else if(tab==='审计日志') setRows(await api('/api/audit'));
      else setRows([]);
    }catch(e:any){setMessage(e.message)}
  };

  useEffect(()=>{load()},[tab]);

  const action=async(fn:()=>Promise<any>)=>{
    setBusy(true);
    try{
      await fn();
      setMessage('操作成功');
      await load();
    }catch(e:any){setMessage(e.message)}
    finally{setBusy(false)}
  };

  const toggleMode=()=>action(()=>api('/api/mode',{
    method:'PUT',
    body:JSON.stringify({mode:dash.configuredMode==='AI_WAREHOUSE_OS'?'TRADITIONAL_WMS':'AI_WAREHOUSE_OS'})
  }));

  const toggleKill=()=>action(()=>api('/api/mode',{
    method:'PUT',
    body:JSON.stringify({killSwitch:!dash.killSwitch})
  }));

  return <div className="shell">
    <aside>
      <div className="brand">
        <div className="brandmark">W</div>
        <div><b>AI Warehouse OS</b><small>Warehouse Brain</small></div>
      </div>

      <div className="warehouse-card">
        <div><strong>华北一号仓</strong><span>ONLINE</span></div>
        <small>ORG-001 · WH-BJ-01</small>
      </div>

      <nav>{tabs.map(t=><button key={t} className={tab===t?'active':''} onClick={()=>setTab(t)}>{t}</button>)}</nav>

      <div className="runtime">
        <small>Runtime</small>
        <b>React + Go</b>
        <span>Vercel Container</span>
      </div>
    </aside>

    <main>
      <header>
        <div><h1>{tab}</h1><p>AI Planning 与确定性 WMS Core 分层运行</p></div>
        <div className="header-actions">
          <div className={'kill '+(dash.killSwitch?'danger':'')} onClick={toggleKill}>
            <span>AI Kill Switch</span><b>{dash.killSwitch?'ON':'OFF'}</b>
          </div>
          <div className="mode-control">
            <span>{dash.effectiveMode==='AI_WAREHOUSE_OS'?'AI Warehouse OS':'Traditional WMS'}</span>
            <button className={'switch '+(dash.configuredMode==='AI_WAREHOUSE_OS'?'on':'')} onClick={toggleMode}><i/></button>
          </div>
        </div>
      </header>

      {message&&<div className="toast" onClick={()=>setMessage('')}>{message}</div>}
      {busy&&<div className="overlay">处理中…</div>}

      {tab==='总览'&&<Overview d={dash}/>}
      {tab==='入库 ASN'&&<ASNPage rows={rows} action={action}/>}
      {tab==='库存'&&<InventoryPage rows={rows}/>}
      {tab==='出库订单'&&<OrdersPage rows={rows} action={action}/>}
      {tab==='作业任务'&&<TaskPage rows={rows}/>}
      {tab==='AI Agent'&&<AIPage enabled={dash.effectiveMode==='AI_WAREHOUSE_OS'} plan={aiPlan} setPlan={setAiPlan} action={action}/>}
      {tab==='审计日志'&&<AuditPage rows={rows}/>}
    </main>
  </div>
}

function Overview({d}:{d:Row}){
  const cards=[
    ['可用库存',d.available||0,'AVAILABLE'],
    ['收货暂存',d.receiving||0,'RECEIVING'],
    ['已分配',d.allocated||0,'ALLOCATED'],
    ['已拣货',d.picked||0,'PICKED'],
    ['出库订单',d.orders||0,'ORDERS'],
    ['待执行任务',d.pendingTasks||0,'TASKS'],
  ];
  return <>
    <div className="metric-grid">{cards.map(([k,v,s])=><div className="metric" key={String(k)}>
      <span>{k}</span><strong>{v}</strong><small>{s}</small>
    </div>)}</div>

    <div className="two-col">
      <section>
        <div className="section-title"><h2>系统运行状态</h2><span className="ok">HEALTHY</span></div>
        <Status name="WMS Core" value="Deterministic"/>
        <Status name="Task Graph" value="Running"/>
        <Status name="WCS Adapter" value="Connected"/>
        <Status name="Inventory Ledger" value="Protected"/>
        <Status name="AI Decision Layer" value={d.effectiveMode==='AI_WAREHOUSE_OS'?'Active':'Standby'} ai={d.effectiveMode==='AI_WAREHOUSE_OS'}/>
      </section>

      <section>
        <div className="section-title"><h2>双模式架构</h2><span>Same Core</span></div>
        <p className="desc">关闭 AI 模式时，系统表现为传统 WMS；打开后，AI Agent 基于同一套订单、库存和任务事实生成计划。无论哪种模式，库存账务与安全约束都由确定性 WMS Core 执行。</p>
        <div className="architecture">
          <div className="layer ai">AI / Optimization</div>
          <div className="arrow">↓</div>
          <div className="layer">Task Graph / Orchestrator</div>
          <div className="arrow">↓</div>
          <div className="layer core">Deterministic WMS Core</div>
        </div>
      </section>
    </div>
  </>
}

function Status({name,value,ai=false}:any){
  return <div className="status-row"><b>{name}</b><span className={ai?'ai-badge':'ok-badge'}>{value}</span></div>
}

function ASNPage({rows,action}:any){
  const [supplier,setSupplier]=useState('可口可乐北京');
  const [sku,setSku]=useState('COKE330');
  const [expected,setExpected]=useState(100);

  return <>
    <section className="form-card">
      <div><h2>创建 ASN</h2><p>预期收货单 → 收货暂存 → PUTAWAY</p></div>
      <input value={supplier} onChange={e=>setSupplier(e.target.value)} placeholder="供应商"/>
      <input value={sku} onChange={e=>setSku(e.target.value)} placeholder="SKU"/>
      <input type="number" value={expected} onChange={e=>setExpected(+e.target.value)} />
      <button className="primary" onClick={()=>action(()=>api('/api/asns',{method:'POST',body:JSON.stringify({supplier,sku,expected})}))}>创建 ASN</button>
    </section>

    <DataTable rows={rows} action={(r:any)=><div className="row-actions">
      <button disabled={r.received>=r.expected} onClick={()=>action(()=>api('/api/asns/'+r.id+'/receive',{method:'POST',body:JSON.stringify({qty:r.expected-r.received})}))}>收货</button>
      <button disabled={r.received<=0||r.status==='PUTAWAY_DONE'} onClick={()=>action(()=>api('/api/asns/'+r.id+'/putaway',{method:'POST'}))}>上架</button>
    </div>}/>
  </>
}

function InventoryPage({rows}:{rows:Row[]}){
  const total=useMemo(()=>rows.reduce((n,r)=>n+(r.available||0)+(r.receiving||0)+(r.allocated||0)+(r.picked||0),0),[rows]);
  return <>
    <div className="summary-bar"><span>Inventory Truth</span><b>{total} units</b><small>AVAILABLE + RECEIVING + ALLOCATED + PICKED</small></div>
    <DataTable rows={rows}/>
  </>
}

function OrdersPage({rows,action}:any){
  const [customer,setCustomer]=useState('北京朝阳门店');
  const [sku,setSku]=useState('COKE330');
  const [qty,setQty]=useState(20);
  return <>
    <section className="form-card">
      <div><h2>创建出库订单</h2><p>Order → Allocation → PICK → PACK → SHIP</p></div>
      <input value={customer} onChange={e=>setCustomer(e.target.value)} placeholder="客户"/>
      <input value={sku} onChange={e=>setSku(e.target.value)} placeholder="SKU"/>
      <input type="number" value={qty} onChange={e=>setQty(+e.target.value)} />
      <button className="primary" onClick={()=>action(()=>api('/api/orders',{method:'POST',body:JSON.stringify({customer,sku,qty})}))}>创建订单</button>
    </section>

    <DataTable rows={rows} action={(r:any)=><div className="row-actions">
      {['allocate','pick','pack','ship'].map(x=><button key={x} onClick={()=>action(()=>api('/api/orders/'+r.id+'/'+x,{method:'POST'}))}>{x.toUpperCase()}</button>)}
    </div>}/>
  </>
}

function TaskPage({rows}:{rows:Row[]}){
  return <>
    <div className="summary-bar"><span>Task Graph</span><b>{rows.length} tasks</b><small>WCS / SYSTEM executor</small></div>
    <DataTable rows={rows}/>
  </>
}

function AIPage({enabled,plan,setPlan,action}:any){
  return <div className="two-col">
    <section>
      <div className="section-title"><h2>Warehouse Agent</h2><span className={enabled?'ok':'muted'}>{enabled?'ACTIVE':'STANDBY'}</span></div>
      <p className="desc">AI 读取库存、订单、ASN 和任务状态，生成动作建议；AI 不直接写 Inventory Ledger。</p>
      {!enabled&&<div className="warning">请先打开右上角 AI Warehouse OS 模式。</div>}
      <button className="primary wide" disabled={!enabled} onClick={()=>action(async()=>setPlan(await api('/api/ai/plan',{method:'POST'})))}>生成当前仓库自主计划</button>
    </section>
    <section>
      <div className="section-title"><h2>Planning Result</h2><span>Read Only</span></div>
      <pre>{plan?JSON.stringify(plan,null,2):'尚未生成计划'}</pre>
    </section>
  </div>
}

function AuditPage({rows}:{rows:Row[]}){
  return <>
    <div className="summary-bar"><span>Append-only Audit View</span><b>{rows.length} events</b><small>actor · action · detail · time</small></div>
    <DataTable rows={rows}/>
  </>
}

function DataTable({rows,action}:{rows:Row[],action?:(r:Row)=>React.ReactNode}){
  if(!rows.length) return <section className="empty">暂无数据，先执行一个业务操作。</section>;
  const keys=Object.keys(rows[0]);
  return <section className="table-card"><table>
    <thead><tr>{keys.map(k=><th key={k}>{k}</th>)}{action&&<th>操作</th>}</tr></thead>
    <tbody>{rows.map((r,i)=><tr key={i}>{keys.map(k=><td key={k}>{String(r[k])}</td>)}{action&&<td>{action(r)}</td>}</tr>)}</tbody>
  </table></section>
}

createRoot(document.getElementById('root')!).render(<App/>);
