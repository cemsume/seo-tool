import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Textarea } from "./components/ui/textarea";
import { Button } from "./components/ui/button";
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
// import { StartCrawl } from "../wailsjs/go/backend/Crawl";
import {
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable"
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { CancelFetch, StartCrawl, SaveFile } from '../wailsjs/go/main/App.js';
import { AgGridReact } from 'ag-grid-react'; // React Data Grid Component
import "ag-grid-community/styles/ag-grid.css"; // Mandatory CSS required by the Data Grid
import "ag-grid-community/styles/ag-theme-quartz.css";
import Drawer from 'react-modern-drawer'
import 'react-modern-drawer/dist/index.css'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime.js'


interface Result {
  Url: string;
  StatusCode: number;
  Type: string;
  Size: number;
  Age: string;
  Redirect: string;
  ExtraData: any;
  Body: string;
  Updated: boolean;
}

function App() {
  const gridRef = useRef<AgGridReact>(null);

  const [urls, setUrls] = useState("")
  const [results, setResults] = useState<Result[]>([])
  const [totalCount, setTotalCount] = useState(0)

  const [isOpen, setIsOpen] = useState(false)

  const toggleDrawer = () => {
    setIsOpen((prevState) => !prevState)
  }

  const handleUrlChange = (event: any) => {
    setUrls(event.target.value);
  };



  const renderBottomPanel = () => {
    return <div className="relative overflow-x-auto max-h-[400px]">
      <table className="w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400">
        <thead className="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
          <tr>
            <th scope="col" className="px-6 py-3">
              Key
            </th>
            <th scope="col" className="px-6 py-3">
              Value
            </th>
          </tr>
        </thead>
        <tbody>
          <tr className="bg-white border-b dark:bg-gray-800 dark:border-gray-700">
            <th scope="row" className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap dark:text-white">
              Preview
            </th>
            <td className="px-6 py-4">
              <button onClick={toggleDrawer}>Show</button>
            </td>
          </tr>
          <tr className="bg-white border-b dark:bg-gray-800 dark:border-gray-700">
            <th scope="row" className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap dark:text-white">
              Url
            </th>
            <td className="px-6 py-4">
              {selectedItem?.Url}
            </td>
          </tr>
          <tr className="bg-white border-b dark:bg-gray-800 dark:border-gray-700">
            <th scope="row" className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap dark:text-white">
              Redirect
            </th>
            <td className="px-6 py-4">
              {selectedItem?.Redirect}
            </td>
          </tr>
          {Object.keys(selectedItem?.ExtraData?.Headers ?? {}).map((key) => (
            <tr className="bg-white border-b dark:bg-gray-800 dark:border-gray-700">
              <th scope="row" className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap dark:text-white">
                {key}
              </th>
              <td className="px-6 py-4">
                {selectedItem?.ExtraData?.Headers[key]}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  }
  const renderDialog = () => {
    return <Dialog>
      <DialogTrigger className="flex cursor-pointer select-none items-center rounded-sm text-sm font-medium outline-none focus:bg-slate-100 focus:text-slate-900">Open</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Paste Urls</DialogTitle>
          <Textarea value={urls}
            onChange={handleUrlChange} className="w-full h-full" rows={35} placeholder="urls" />
        </DialogHeader>
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="default" onClick={onClickUrlButton}>
              Ok
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  }

  const renderNavMenu = () => {
    return <div className="pl-20 flex space-x-4  h-10 content-between rounded-md border border-slate-200 bg-white p-1 dark:border-slate-800 dark:bg-slate-950" >
      {renderDialog()}
      {renderUserAgentRadioGroup()}
      <div>
        {results.length} / {totalCount}
      </div>
      <div>
        v0.0.31.1
      </div>
      <Button onClick={() => SaveFile(results)}> Export </Button>
    </div>
  }


  const [lastChannel, setLastChannel] = useState("")
  const onClickUrlButton = () => {
    setResults([])
    CancelFetch();
    setSelectedItem(undefined)
    const newResults = urls.split("\n").map((url) => ({
      Url: url,
      StatusCode: 0,
      Type: "",
      Size: 0,
      Age: "",
      Redirect: "",
      ExtraData: {},
    })) as Result[];
    setTotalCount(newResults.length)
    StartCrawl(urls, selectedValue).then((response) => {
      setUrls("")
      gridRef.current!.api.applyTransactionAsync({ remove: results })
      if (lastChannel !== response && lastChannel !== "") {
        EventsOff(lastChannel)
      }
      setLastChannel(response)

      EventsOn(response, (result: Result) => {
        setTimeout(() => {
          gridRef.current!.api.applyTransactionAsync(
            { add: [result] },
          );
        }, 0);

        setResults((prevResults) => [...prevResults, result]);
      });
    })
  }


  const [selectedValue, setSelectedValue] = useState('desktop');

  const handleChange = (event: any) => {
    setSelectedValue(event);
  };

  const renderUserAgentRadioGroup = () => {
    return <RadioGroup onValueChange={handleChange} className="w-[600px] grid-flow-col pl-20" title="User-Agent" orientation="horizontal" defaultValue="option-one">
      <div className="flex items-center space-x-2">
        <RadioGroupItem value="desktop" id="desktop" />
        <Label htmlFor="desktop">Desktop</Label>
      </div>
      <div className="flex items-center space-x-2">
        <RadioGroupItem value="mobile" id="mobile"
        />
        <Label htmlFor="mobile">Mobile</Label>
      </div>
      <div className="flex items-center space-x-2">
        <RadioGroupItem value="bot-desktop" id="bot-desktop"
        />
        <Label htmlFor="bot-desktop">Bot Desktop</Label>
      </div>
      <div className="flex items-center space-x-2">
        <RadioGroupItem value="bot-mobile" id="bot-mobile" />
        <Label htmlFor="bot-mobile">Bot Mobile</Label>
      </div>
    </RadioGroup>
  }

  const [selectedItem, setSelectedItem] = useState<Result>()

  const onRowSelected = (event: any) => {
    if (event.node.selected) {
      setSelectedItem(event.data);
    }
  };



  const onBtnExport = () => {
    console.log("Exporting data");
    const params = getParams();
    gridRef.current!.api.exportDataAsCsv(params);
  };

  function getParams() {
    return {
      columnSeparator: ';',
      suppressQuotes: true
    };
  }
  return (
    <>
      {renderNavMenu()}
      <Drawer
        open={isOpen}
        onClose={toggleDrawer}
        direction='right'
        size={1550}
        className="h-max overflow-auto"
        lockBackgroundScroll={true}
      >
        <div className="h-max overflow-auto"
          dangerouslySetInnerHTML={{
            __html: selectedItem?.Body ?? ''
          }}></div>
      </Drawer>
      <div className="min-h-screen grid justify-items-stretch">
        <ResizablePanelGroup direction="vertical">
          <ResizablePanel>
            <div className="ag-theme-quartz" style={{ height: 500 }}>
              <AgGridReact
                ref={gridRef}
                suppressExcelExport={true}
                onRowSelected={onRowSelected}
                rowSelection="single"
                loading={false}
                columnDefs={
                  [
                    { headerName: "URL", field: "Url", width: 400, sortable: true, filter: true, flex: 1, minWidth: 550 },
                    { headerName: "Status Code", field: "StatusCode", sortable: true, filter: true, flex: 1 },
                    { headerName: "Type", field: "Type", sortable: true, filter: true, flex: 1 },
                    { headerName: "Size", field: "Size", sortable: true, filter: true, flex: 1 },
                    { headerName: "Age", field: "Age", sortable: true, filter: true, flex: 1 },
                    { headerName: "Redirect", field: "Redirect", sortable: true, filter: true, flex: 1 },
                  ]
                } />
            </div>
          </ResizablePanel>
          <ResizablePanel>
            {selectedItem && renderBottomPanel()}
          </ResizablePanel>
        </ResizablePanelGroup>


      </div>
    </>
  )
}

export default App
