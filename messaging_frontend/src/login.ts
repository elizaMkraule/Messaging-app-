/**
 * Class LoginPage is a custom webcomponent for the login page
 *
 * handles the logging in a user by collecting the name submited by user and dispatching a custom event to handle the authorization
 *
 * @extends HTMLElement
 */
export class LoginPage extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });

    if (this.shadowRoot) {
      this.shadowRoot.innerHTML = `
        <style>
            /* Add CSS styles here? or does it apply from the syle sheet? */
        </style>
        <dialog id="loginModal">
            <form id="loginForm">
                <input type="text" id="username" placeholder="Enter your username" />
                <button type="submit" id="textform">Login</button>
                <p id="error-message" style="color: red; display: none;">Oops! It looks like you forgot to enter your username. Please fill it in to log in.</p>
                <p id="error-message-db" style="color: red; display: none;">Error: Could not connect to database.</p>
            </form>
            </form>
        </dialog>
    `;
    }
  }

  connectedCallback() {
    const loginModal = this.shadowRoot?.getElementById(
      "loginModal",
    ) as HTMLDialogElement;
    const usernameInput = this.shadowRoot?.getElementById(
      "username",
    ) as HTMLInputElement;
    const textform = this.shadowRoot?.getElementById(
      "textform",
    ) as HTMLInputElement;

    const errorMessage = this.shadowRoot?.getElementById("error-message");

    loginModal.showModal();

    const newForm = document.createElement("form");
    newForm.id = "loginForm";
    const submit = document.createElement("button");
    submit.textContent = "Login";
    newForm.append(submit);

    // Handler for clicking on the login button
    submit.addEventListener("click", (event: MouseEvent) => {
      console.log("Clicked submit button");
      event.preventDefault();
    });

    // Handler for submiting the user name
    textform.addEventListener("click", (event: MouseEvent) => {
      event.preventDefault();
      console.log(usernameInput.value);

      const loginEvent = new CustomEvent("loginEvent", {
        detail: { message: usernameInput.value },
      });
      if (usernameInput.value.trim() === "") {
        if (errorMessage) {
          errorMessage.style.display = "block";
          errorMessage.setAttribute("aria-live", "assertive");
        }
      } else {
        if (errorMessage) {
          errorMessage.style.display = "none";
          // Notification of login event
          document.dispatchEvent(loginEvent);
        }
      }
    });
  }

  displayError(): void {
    const errorMessage = this.shadowRoot?.getElementById("error-message-db");
    if (errorMessage) {
      errorMessage.style.display = "block";
      errorMessage.setAttribute("aria-live", "assertive");
    }
  }

  // Remove event listeners to avoid memory leaks
  disconnectedCallback() {
    const loginForm = this.shadowRoot?.getElementById("loginForm");
    if (loginForm) {
      loginForm.remove();
    }
  }
}
