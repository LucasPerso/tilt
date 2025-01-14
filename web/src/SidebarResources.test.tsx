import { mount, ReactWrapper } from "enzyme"
import React from "react"
import { MemoryRouter } from "react-router"
import {
  cleanupMockAnalyticsCalls,
  expectIncrs,
  mockAnalyticsCalls,
} from "./analytics_test_helpers"
import { accessorsForTesting, tiltfileKeyContext } from "./LocalStorage"
import LogStore from "./LogStore"
import {
  AlertsOnTopToggle,
  ResourceNameFilterTextField,
} from "./OverviewSidebarOptions"
import { assertSidebarItemsAndOptions } from "./OverviewSidebarOptions.test"
import PathBuilder from "./PathBuilder"
import SidebarItem from "./SidebarItem"
import SidebarResources from "./SidebarResources"
import { StarredResourcesContextProvider } from "./StarredResourcesContext"
import StarResourceButton from "./StarResourceButton"
import {
  oneResource,
  oneResourceTestWithName,
  twoResourceView,
} from "./testdata"
import { ResourceView, SidebarOptions } from "./types"

let pathBuilder = PathBuilder.forTesting("localhost", "/")

const sidebarOptionsAccessor = accessorsForTesting<SidebarOptions>(
  "sidebar_options"
)
const starredItemsAccessor = accessorsForTesting<string[]>("pinned-resources")

function clickStar(
  root: ReactWrapper<any, React.Component["state"], React.Component>,
  name: string
) {
  let starButtons = root.find(StarResourceButton).find({ resourceName: name })
  expect(starButtons.length).toBeGreaterThan(0)
  starButtons.at(0).simulate("click")
}

describe("SidebarResources", () => {
  beforeEach(() => {
    mockAnalyticsCalls()
  })

  afterEach(() => {
    cleanupMockAnalyticsCalls()
    localStorage.clear()
  })

  it("adds items to the starred list when items are starred", () => {
    let ls = new LogStore()
    let items = twoResourceView().uiResources.map((r) => new SidebarItem(r, ls))
    const root = mount(
      <MemoryRouter>
        <tiltfileKeyContext.Provider value="test">
          <StarredResourcesContextProvider>
            <SidebarResources
              items={items}
              selected={""}
              resourceView={ResourceView.Log}
              pathBuilder={pathBuilder}
            />
          </StarredResourcesContextProvider>
        </tiltfileKeyContext.Provider>
      </MemoryRouter>
    )

    clickStar(root, "snack")

    expectIncrs(
      { name: "ui.web.star", tags: { starCount: "0", action: "load" } },
      {
        name: "ui.web.sidebarStarButton",
        tags: { action: "click", newStarState: "true" },
      },
      { name: "ui.web.star", tags: { starCount: "1", action: "star" } }
    )

    expect(starredItemsAccessor.get()).toEqual(["snack"])
  })

  it("removes items from the starred list when items are unstarred", () => {
    let ls = new LogStore()
    let items = twoResourceView().uiResources.map((r) => new SidebarItem(r, ls))
    starredItemsAccessor.set(items.map((i) => i.name))

    const root = mount(
      <MemoryRouter>
        <tiltfileKeyContext.Provider value="test">
          <StarredResourcesContextProvider>
            <SidebarResources
              items={items}
              selected={""}
              resourceView={ResourceView.Log}
              pathBuilder={pathBuilder}
            />
          </StarredResourcesContextProvider>
        </tiltfileKeyContext.Provider>
      </MemoryRouter>
    )

    clickStar(root, "snack")

    expectIncrs(
      { name: "ui.web.star", tags: { starCount: "2", action: "load" } },
      {
        name: "ui.web.sidebarStarButton",
        tags: { action: "click", newStarState: "false" },
      },
      { name: "ui.web.star", tags: { starCount: "1", action: "unstar" } }
    )

    expect(starredItemsAccessor.get()).toEqual(["vigoda"])
  })

  const falseyOptions: SidebarOptions = {
    alertsOnTop: false,
    resourceNameFilter: "",
  }

  const loadCases: [string, any, string[]][] = [
    [
      "alertsOnTop",
      { ...falseyOptions, alertsOnTop: true },
      ["vigoda", "a", "b"],
    ],
    [
      "resourceNameFilter",
      { ...falseyOptions, resourceNameFilter: "vig" },
      ["vigoda"],
    ],
    [
      "resourceNameFilter undefined",
      { ...falseyOptions, resourceNameFilter: undefined },
      ["vigoda", "a", "b"],
    ],
  ]
  test.each(loadCases)(
    "loads %p from localStorage",
    (name, options, expectedItems) => {
      sidebarOptionsAccessor.set(options)

      let ls = new LogStore()
      const items = [
        oneResource(),
        oneResourceTestWithName("a"),
        oneResourceTestWithName("b"),
      ].map((res) => new SidebarItem(res, ls))

      const root = mount(
        <MemoryRouter>
          <tiltfileKeyContext.Provider value="test">
            <SidebarResources
              items={items}
              selected={""}
              resourceView={ResourceView.OverviewDetail}
              pathBuilder={pathBuilder}
            />
          </tiltfileKeyContext.Provider>
        </MemoryRouter>
      )

      assertSidebarItemsAndOptions(root, expectedItems, options.alertsOnTop)
    }
  )

  const saveCases: [string, SidebarOptions][] = [
    ["alertsOnTop", { ...falseyOptions, alertsOnTop: true }],
    ["resourceNameFilter", { ...falseyOptions, resourceNameFilter: "foo" }],
  ]
  test.each(saveCases)(
    "saves option %s to localStorage",
    (name, expectedOptions) => {
      let ls = new LogStore()
      const items = [
        oneResource(),
        oneResourceTestWithName("a"),
        oneResourceTestWithName("b"),
      ].map((res) => new SidebarItem(res, ls))

      const root = mount(
        <MemoryRouter>
          <tiltfileKeyContext.Provider value="test">
            <SidebarResources
              items={items}
              selected={""}
              resourceView={ResourceView.OverviewDetail}
              pathBuilder={pathBuilder}
            />
          </tiltfileKeyContext.Provider>
        </MemoryRouter>
      )

      let aotToggle = root.find(AlertsOnTopToggle)
      if (aotToggle.hasClass("is-enabled") !== expectedOptions.alertsOnTop) {
        aotToggle.simulate("click")
      }

      let resourceNameFilterTextField = root.find(ResourceNameFilterTextField)
      if (
        resourceNameFilterTextField.props().value !==
        expectedOptions.resourceNameFilter
      ) {
        resourceNameFilterTextField.find("input").simulate("change", {
          target: { value: expectedOptions.resourceNameFilter },
        })
      }

      const observedOptions = sidebarOptionsAccessor.get()
      expect(observedOptions).toEqual(expectedOptions)
    }
  )
})
