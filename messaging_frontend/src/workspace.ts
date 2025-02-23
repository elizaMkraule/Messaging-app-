import { ViewWorkspace } from "./types";

/**
 * Class WorkspacePage is a custom webcomponent for dispalying the page setup
 *
 *
 * @extends HTMLElement
 */
export class WorkspacePage extends HTMLElement {
  channelsSection: any;
  postsSection: any;
  workspaceSection: any;
  textArea: any;
  messageForm: any;
  createWorkspace: any;
  newWorkspace: any;
  postlist: any;
  observer: any;

  constructor() {
    super();
    this.attachShadow({ mode: "open" });

    const template = document.getElementById(
      "WorkspaceLayout",
    ) as HTMLTemplateElement;
    // Clone the content of the template
    const clone = document.importNode(template.content, true);
    const shadowRoot = this.shadowRoot;
    shadowRoot?.appendChild(clone);

    this.postlist = this.shadowRoot?.querySelector("#posts-list");

    this.postsSection = this.shadowRoot?.querySelector("#posts");
    this.channelsSection = this.shadowRoot?.querySelector("#channels");
    this.workspaceSection = this.shadowRoot?.querySelector("#workspaces");
    this.messageForm = this.shadowRoot?.querySelector("#new-message-form");
    this.createWorkspace = this.shadowRoot?.querySelector(
      "#CreateWorkspaceForm",
    );
    this.newWorkspace = this.shadowRoot?.querySelector(
      "#usernameWorkspace",
    ) as HTMLInputElement;
  }

  /**
   * Displays an error message depedning on the error code in the workspace
   *
   * @param errorCode the code from the error
   */
  displayWorkspaceError(errorCode: string): void {
    this.removeWorkspaceErrors();
    console.log("called dispaly error in workspaces");
    console.log("got error", errorCode);
    let errorMessage = "";
    if (errorCode == "Error: 400") {
      errorMessage = "Workspace already exists.";
    } else if (errorCode == "Error: 404") {
      errorMessage = "Missing resource: Workspace does not exist.";
    } else {
      errorMessage = "An error occurred:" + errorCode;
    }

    const errorSection = this.createErrorElement(errorMessage);

    // Append the errorElement to the shadow DOM
    this.workspaceSection.appendChild(errorSection);
  }

  /**
   * Displays an error message depedning on the error code in the channel
   *
   * @param errorCode the code from the error
   */

  displayChannelError(errorCode: string): void {
    this.removeChannelErrors();
    console.log("called dispaly error in channels");
    console.log("got error", errorCode);

    let errorMessage = "";
    if (errorCode == "Error: 400") {
      errorMessage = "Channel already exists.";
    } else if (errorCode == "Error: 404") {
      errorMessage = "Channel does not exist.";
    } else {
      errorMessage = "An unknown error occurred.";
    }

    const errorSection = this.createErrorElement(errorMessage);
    // Append the errorElement to the shadow DOM
    this.channelsSection.appendChild(errorSection);
  }

  /**
   * Removes all channel errors
   *
   */
  removeChannelErrors(): void {
    const error = this.channelsSection?.querySelector(`[id = "error"]`);
    if (error) {
      error.remove();
    }
  }

  /**
   * Removes all workspace errors
   *
   */
  removeWorkspaceErrors(): void {
    const error = this.workspaceSection?.querySelector(`[id = "error"]`);
    if (error) {
      error.remove();
    }
  }

  /**
   * Setting up event listeners for workspace creation
   */
  connectedCallback(): void {
    this.createWorkspace?.addEventListener("submit", (event: Event) => {
      event.preventDefault();
      console.log(this.newWorkspace?.value);

      const createWorkspaceEvent = new CustomEvent("createWorkspaceEvent", {
        detail: { workspace: this.newWorkspace?.value },
      });
      console.log(createWorkspaceEvent);
      // Notification of login event.
      document.dispatchEvent(createWorkspaceEvent);
      this.newWorkspace.value = "";
    });
  }

  disconnectedCallback(): void {}

  /**
   * Creates an error element to be displayed
   *
   * @param errorMessage a string with the message to be displayed as an error
   */
  createErrorElement(errorMessage: string): HTMLElement {
    // Create error display container
    const errorSection = document.createElement("section");

    errorSection.setAttribute("id", "error");
    errorSection.style.cssText = `
        color: red;
        padding: 1vh;
        margin: 1vhpx 0;
        border-radius: 5px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        background-color: white;
        border: 1px solid red;
    `;

    // Create the paragraph for the error message
    const errorMessageP = document.createElement("p");
    errorMessageP.textContent = errorMessage;
    errorMessageP.setAttribute("aria-live", "assertive");
    errorMessageP.style.fontSize = "2vh";

    // Create dismiss button
    const dismiss = document.createElement("button");
    dismiss.textContent = "Dismiss error";
    dismiss.className = "dismissButton";
    dismiss.setAttribute("aria-label", "Dismiss error");

    // Add click event to dismiss the error message
    dismiss.addEventListener("click", () => {
      errorSection.remove();
    });

    // Append the message and button to the errorSection
    errorSection.appendChild(errorMessageP);
    errorSection.appendChild(dismiss);

    return errorSection;
  }
}

/**
 * Class WorkspaceItem is a custom webcomponent a workspaceItem
 *
 * @extends HTMLElement
 */
export class WorkspaceItem extends HTMLElement {
  private _data: ViewWorkspace | null = null;
  workspaceButton: any;

  constructor() {
    super();
    // console.log("constructing");
    this.attachShadow({ mode: "open" });

    const template = document.getElementById(
      "WorkspaceItemTemplate",
    ) as HTMLTemplateElement;
    // Clone the content of the template --> not sure if doing this correctly
    const clone = document.importNode(template.content, true);
    const shadowRoot = this.shadowRoot;
    shadowRoot?.appendChild(clone);

    this.workspaceButton = this.shadowRoot?.querySelector("#wrkspcButton");
  }

  set data(data: ViewWorkspace) {
    this._data = data;
  }

  connectedCallback(): void {
    this.display();

    document.addEventListener("workspaceDeleted", (event: CustomEvent) => {
      if (this._data?.path === event.detail.id) {
        this.remove(); // This removes the component from the DOM
      }
    });
  }

  disconnectedCallback(): void {}

  // Method to active state to change color of item
  ActiveState(): void {
    this.workspaceButton.classList.add("active");
  }
  // Method to deactive state to change color of item
  DeactivateState(): void {
    this.workspaceButton.classList.remove("active");
  }

  private display(): void {
    console.log(this._data);
    if (this.shadowRoot) {
      const path = JSON.stringify(this._data?.path, null, 2);

      const pathstring = path.replace(/^"\/|"$|^\//g, "");

      this.workspaceButton.textContent = pathstring || "none";

      const item = document.createElement("p");
      const label = document.createElement("label");

      // Create open button
      const openIcon = document.createElement("iconify-icon");
      openIcon.setAttribute("icon", "bi:box-arrow-in-right");
      openIcon.setAttribute("width", "1.25em");
      openIcon.setAttribute("height", "1.25em");
      const open = document.createElement("button");
      open.append(openIcon);
      open.setAttribute("role", "button");
      open.setAttribute("aria-label", "open workspace");

      open.className = "workspace-small-btn";

      //  add an event to the open button that we want to open this workspace
      open.addEventListener("click", () => {
        console.log("workspace open button clicked");
        const openEvent = new CustomEvent("openEvent", {
          detail: { name: this._data?.path },
        });
        document.dispatchEvent(openEvent);
        this.ActiveState();
      });

      // Create delete button
      const trashIcon = document.createElement("iconify-icon");
      trashIcon.setAttribute("icon", "iconamoon:trash-fill");
      trashIcon.setAttribute("width", "1.25em");
      trashIcon.setAttribute("height", "1.25em");
      const trash = document.createElement("button");
      trash.append(trashIcon);
      trash.setAttribute("aria-label", "delete workspace");
      trash.className = "workspace-small-btn";

      //   Delete handler
      trash.addEventListener("click", () => {
        const DeleteWorkspace = new CustomEvent("deleteWorkspace", {
          detail: { id: this._data?.path },
        });

        console.log("clicked on deleting a workspace");
        document.dispatchEvent(DeleteWorkspace);
      });

      label.append(path.replace(/^"\/|"$|^\//g, ""));
      item.append(trash, open);

      this.workspaceButton?.appendChild(item);
    } else {
      throw new Error("shadowRoot does not exist");
    }
  }
}
